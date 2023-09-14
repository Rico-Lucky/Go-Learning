# net/http

## 1.CS架构示例

服务端代码：

```go
package main

import "net/http"

func main() {

    // 注册对应于请求路径 /ping 的handler函数
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})  

   //启动一个端口为8080的http服务
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

```

客户端单元测试代码：

```sh
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func Client() {

	rsp, err := http.Post("http://localhost:8080/ping", "", nil)
	if err != nil {
		panic(err)
	}
	str, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()
	fmt.Println(string(str))

}

func TestClient(t *testing.T) {
	Client()
}
```

## 2.服务端

### 2.1 核心数据结构

- （1）Server

  ```go
  type Server struct {
  	// Addr optionally specifies the TCP address for the server to listen on,
  	// in the form "host:port". If empty, ":http" (port 80) is used.
  	// The service names are defined in RFC 6335 and assigned by IANA.
  	// See net.Dial for details of the address format.
  	Addr string      //服务地址
  
  	Handler Handler // 想当于路由处理器。实现从请求路径path到具体处理函数handler的注册和映射能力。在用户构造Server对象时，若其中的Handler字段未显示声明，则会取net/http包下单例对象DefaultServerMux(ServerMux类型)进行兜底
  
  	// DisableGeneralOptionsHandler, if true, passes "OPTIONS *" requests to the Handler,
  	// otherwise responds with 200 OK and Content-Length: 0.
  	DisableGeneralOptionsHandler bool
  
  	// TLSConfig optionally provides a TLS configuration for use
  	// by ServeTLS and ListenAndServeTLS. Note that this value is
  	// cloned by ServeTLS and ListenAndServeTLS, so it's not
  	// possible to modify the configuration with methods like
  	// tls.Config.SetSessionTicketKeys. To use
  	// SetSessionTicketKeys, use Server.Serve with a TLS Listener
  	// instead.
  	TLSConfig *tls.Config
  
  	// ReadTimeout is the maximum duration for reading the entire
  	// request, including the body. A zero or negative value means
  	// there will be no timeout.
  	//
  	// Because ReadTimeout does not let Handlers make per-request
  	// decisions on each request body's acceptable deadline or
  	// upload rate, most users will prefer to use
  	// ReadHeaderTimeout. It is valid to use them both.
  	ReadTimeout time.Duration
  
  	// ReadHeaderTimeout is the amount of time allowed to read
  	// request headers. The connection's read deadline is reset
  	// after reading the headers and the Handler can decide what
  	// is considered too slow for the body. If ReadHeaderTimeout
  	// is zero, the value of ReadTimeout is used. If both are
  	// zero, there is no timeout.
  	ReadHeaderTimeout time.Duration
  
  	// WriteTimeout is the maximum duration before timing out
  	// writes of the response. It is reset whenever a new
  	// request's header is read. Like ReadTimeout, it does not
  	// let Handlers make decisions on a per-request basis.
  	// A zero or negative value means there will be no timeout.
  	WriteTimeout time.Duration
  
  	// IdleTimeout is the maximum amount of time to wait for the
  	// next request when keep-alives are enabled. If IdleTimeout
  	// is zero, the value of ReadTimeout is used. If both are
  	// zero, there is no timeout.
  	IdleTimeout time.Duration
  
  	// MaxHeaderBytes controls the maximum number of bytes the
  	// server will read parsing the request header's keys and
  	// values, including the request line. It does not limit the
  	// size of the request body.
  	// If zero, DefaultMaxHeaderBytes is used.
  	MaxHeaderBytes int
  
  	// TLSNextProto optionally specifies a function to take over
  	// ownership of the provided TLS connection when an ALPN
  	// protocol upgrade has occurred. The map key is the protocol
  	// name negotiated. The Handler argument should be used to
  	// handle HTTP requests and will initialize the Request's TLS
  	// and RemoteAddr if not already set. The connection is
  	// automatically closed when the function returns.
  	// If TLSNextProto is not nil, HTTP/2 support is not enabled
  	// automatically.
  	TLSNextProto map[string]func(*Server, *tls.Conn, Handler)
  
  	// ConnState specifies an optional callback function that is
  	// called when a client connection changes state. See the
  	// ConnState type and associated constants for details.
  	ConnState func(net.Conn, ConnState)
  
  	// ErrorLog specifies an optional logger for errors accepting
  	// connections, unexpected behavior from handlers, and
  	// underlying FileSystem errors.
  	// If nil, logging is done via the log package's standard logger.
  	ErrorLog *log.Logger
  
  	// BaseContext optionally specifies a function that returns
  	// the base context for incoming requests on this server.
  	// The provided Listener is the specific Listener that's
  	// about to start accepting requests.
  	// If BaseContext is nil, the default is context.Background().
  	// If non-nil, it must return a non-nil context.
  	BaseContext func(net.Listener) context.Context
  
  	// ConnContext optionally specifies a function that modifies
  	// the context used for a new connection c. The provided ctx
  	// is derived from the base context and has a ServerContextKey
  	// value.
  	ConnContext func(ctx context.Context, c net.Conn) context.Context
  
  	inShutdown atomic.Bool // true when server is in shutdown
  
  	disableKeepAlives atomic.Bool
  	nextProtoOnce     sync.Once // guards setupHTTP2_* init
  	nextProtoErr      error     // result of http2.ConfigureServer if used
  
  	mu         sync.Mutex
  	listeners  map[*net.Listener]struct{}
  	activeConn map[*conn]struct{}
  	onShutdown []func()
  
  	listenerGroup sync.WaitGroup
  }
  ```



- （2）Handler 

  路由处理器，根据http请求Request中的请求路径path映射到对应的handler处理函数，对请求进行处理和响应。

  ```go
  type Handler interface {
  	ServeHTTP(ResponseWriter, *Request)
  }
  ```

- （3）ServerMux 对Handler 的具体实现，内部通过一个map维护从path到handler的映射关系

  ```go
  type ServeMux struct {
  	mu    sync.RWMutex
  	m     map[string]muxEntry
  	es    []muxEntry // slice of entries sorted from longest to shortest.
  	hosts bool       // whether any patterns contain hostnames
  }
  ```

- （4）muxEntry 作为一个handler单元，内部包含了请求路径 path + 处理函数handler两部分

  ```go
  type muxEntry struct {
  	h       Handler
  	pattern string
  }
  ```

### 2.2 注册 handler

![](./tmp/go-http-handler注册主干链路.PNG)

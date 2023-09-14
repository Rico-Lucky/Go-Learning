package main

import (
	"net/http"
	"runtime/pprof"
)

var quit chan struct{} = make(chan struct{})

func f() {
	<-quit //读缓冲通道
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

// 这例子中，我们启动了10000个goroutine，并阻塞，然后通过访问http://localhost:11181/，我们就可以得到整个goroutine的信息，仅列出关键信息

// $ curl http://localhost:11181
// goroutine profile: total 10003
// 10000 @ 0xa3251d 0x9f71de 0x9f6e98 0xd85685 0xa5fb81
// #       0xd85684        main.f+0x24     d:/GoPath/go/src/test/pprof/main.go:11
func main() {
	for i := 0; i < 10000; i++ {
		go f()
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":11181", nil)
}

// 造成goroutine泄露的几个原因：

// 1. 从 channel 里读，但是同时没有写入操作
// 2. 向 无缓冲 channel 里写，但是同时没有读操作
// 3. 向已满的 有缓冲 channel 里写，但是同时没有读操作
// 4. select操作在所有case上都阻塞()
// 5. goroutine进入死循环，一直结束不了
// 可见，很多都是因为channel使用不当造成阻塞，从而导致goroutine也一直阻塞无法退出导致的。

//--------------------------------------
// 检测goroutine泄露
//可以使用pprof做分析，但大多数情况都是发生在事后，无法在开发阶段就把问题提早暴露(即“测试左移”)

//------------------------------------------------
// goroutine终止的场景
// 一个goroutine终止有以下几种情况：

// 当一个goroutine完成它的工作
// 由于发生了没有处理的错误
// 有其他的协程告诉它终止
//-------------------------

// 总结-------------------------------------------------
// goroutine leak往往是由于协程在channel上发生阻塞，或协程进入死循环，特别是在一些后台的常驻服务中。
// 在使用channel和goroutine时要注意：

// 创建goroutine时就要想好该goroutine该如何结束
// 使用channel时，要考虑到channel阻塞时协程可能的行为
// 要注意平时一些常见的goroutine leak的场景，包括：master-worker模式，producer-consumer模式等等 。

// 死锁(dead lock):
// 同一个goroutine中，使用同一个chnnel读写；
// 2个 以上的go程中， 使用同一个 channel 通信。 读写channel 先于 go程创建；
// channel 和 读写锁、互斥锁混用；

// 无限死循环(infinite loops)
// I/O 操作上的堵塞也可能造成泄露，例如发送请求到 API 服务器，而没有使用超时；或者程序单纯地陷入死循环中。

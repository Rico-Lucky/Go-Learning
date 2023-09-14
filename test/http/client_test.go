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

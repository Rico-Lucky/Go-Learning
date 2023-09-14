package main

import (
	"fmt"
	"time"
)

// 可能导致goroutine 泄露（goroutine 并未按预期退出并销毁）。由于 select 命中了超时逻辑，导致通道没有消费者（无接收操作），而其定义的通道为无缓冲通道，因此 goroutine 中的ch <- "job result"操作会一直阻塞，最终导致 goroutine 泄露。
func main() {
	ch := make(chan string)
	go func() {
		// 这里假设执行一些耗时的操作
		time.Sleep(15 * time.Second)
		ch <- "job result" //
	}()

	select {
	case result := <-ch:
		fmt.Println(result)
	case <-time.After(time.Second): // 较小的超时时间 可能选中这个case,main函数结束了,最终导致goroutin泄露：
		return
	}
}

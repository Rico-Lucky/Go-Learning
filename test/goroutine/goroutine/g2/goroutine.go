package main

import (
	"fmt"
	"sync"
)

var wait sync.WaitGroup //结构体

func main() {

	for i := 0; i < 100; i++ {
		wait.Add(1) //计数+1
		go func(i int) {
			fmt.Println("你好:", i)
			wait.Done() //通知wait把计数器减一
		}(i)

	}

	fmt.Print("hello main")

	// time.Sleep(1)  //1 是纳秒
	// time.Sleep(time.Second) //1 s

	wait.Wait() //等待所有goroutine任务执行完
}

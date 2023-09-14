package main

import (
	"fmt"
	"sync"
)

var wg sync.WaitGroup

func main() {
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			fmt.Println("A:", i) //闭包，引用外部变量i, 5个[0,5]随机数，有可能重复
			wg.Done()
		}()
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			fmt.Println("B:", i) //闭包，引用外部变量i, 0~4随机顺序
			wg.Done()
		}(i)
	}

	wg.Wait()

}

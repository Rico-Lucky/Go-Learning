package main

import (
	"fmt"
	"runtime"
	"sync"
)

var wg sync.WaitGroup

func a() {
	for i := 1; i <= 10; i++ {
		fmt.Println("A:", i)
	}
	wg.Done()
}

func b() {
	for i := 1; i <= 10; i++ {
		fmt.Println("B:", i)
	}
	wg.Done()
}

func main() {
	runtime.GOMAXPROCS(1) //只占1个CPU内核, 先执行完a或b
	// runtime.GOMAXPROCS(2) //顺序混合执行
	go a()
	go b()
	wg.Wait()

}

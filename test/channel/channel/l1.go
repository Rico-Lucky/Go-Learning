package main

import "fmt"

// 通过通信共享内存，而非共享内存实现通信
func main() {

	// var ch chan int // 声明 var 变量名 chan 元素类型  未初始化的通道类型变量为默认零值是nil 。引用类型，需要初始化后才能使用
	// fmt.Println(ch) // nil

	//通道初始化，使用make内建函数初始化。

	//声明并初始化
	ch2 := make(chan int) //无缓冲通道,同步通道
	// ch2 := make(chan int, 10) //有缓冲通道，10为环形数组的容量

	ch2 <- 10 //若往无缓冲通道写入数据,会阻塞，若无另外的goroutine读取无缓冲通道，那么程序会一直被挂起，直到报错：fatal error: all goroutines are asleep - deadlock! （解决死锁办法可以使用有缓冲区的通道）
	x := <-ch2
	fmt.Println(x)
	// close(ch2)

	/////////////////////////////////////////////////
}

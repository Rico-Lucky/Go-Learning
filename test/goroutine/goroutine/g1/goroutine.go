package main

import (
	"fmt"
	"strings"
	"sync"
)

func main() {
	// goroutine 练习
	// 	交替打印数字和字⺟
	// 问题描述
	// 使⽤两个  goroutine 交替打印序列，⼀个  goroutine 打印数字， 另外⼀
	// 个  goroutine 打印字⺟， 最终效果如下：
	// 12AB34CD56EF78GH910IJ1112KL1314MN1516OP1718QR1920ST2122UV2324WX2526YZ2728

	letter, number := make(chan bool), make(chan bool)

	wait := sync.WaitGroup{}

	// 打印数字
	go func() {
		i := 1
		for {
			select {
			case <-number:
				fmt.Print(i)
				i++
				fmt.Print(i)
				i++
				letter <- true
				break
			default:
				break
			}
		}

	}()

	wait.Add(1)

	// 打印字母
	go func(wait *sync.WaitGroup) {
		str := "ABCDEFGHIJKLMNOPQRSTUVWSYZ"
		i := 0
		for {
			select {
			case <-letter:
				//判断结束标志
				if i >= strings.Count(str, "")-1 {
					wait.Done()
					return
				}

				fmt.Print(str[i : i+1]) // go 的内置切片语法截取字符串,这是按字节截取，在处理 ASCII 单字节字符串截取,没有什么比这更完美的方案了.
				i++

				if i >= strings.Count(str, "") {
					i = 0
				}

				fmt.Print(str[i : i+1])
				i++
				number <- true
				break
			default:
				break
			}
		}
	}(&wait)

	//先打印数字
	number <- true

	//等待所有goroutine 结束
	wait.Wait()
}

package main

//两个goroutine，两个channel
// 1.生产0-100 个数字发送到ch1
// 2. 从ch1中取出数据计算他们的平方，把结果发送到ch2 中

//单向通道  chan<-  只能写
//单向通道  <-chan  只能读

func f1(ch chan<- int) {
	for i := 0; i <= 100; i++ {
		ch <- i
	}
	close(ch)
}

func f2(ch1 <-chan int, ch2 chan<- int) {
	//从通道中取值方式1
	for {
		if tmp, ok := <-ch1; ok {
			ch2 <- tmp * tmp
		} else {
			break
		}
	}
	close(ch2)
}

func main() {
	ch1, ch2 := make(chan int, 100), make(chan int, 200) //带缓冲
	go f1(ch1)
	go f2(ch1, ch2)

	//从通道中取值方式2
	for ret := range ch2 {
		println(ret)
	}
}

// 死锁(dead lock)
// 同一个goroutine中，使用同一个chnnel读写；
// 2个 以上的go程中， 使用同一个 channel 通信。 读写channel 先于 go程创建；
// channel 和 读写锁、互斥锁混用；

func demo3() {
	ch := make(chan int)
	ch <- 1 //这里一直阻塞，运行不到下面
	<-ch
}





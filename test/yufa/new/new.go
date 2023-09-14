package main

import "fmt"

type Param map[string]any
type Show struct {
	Param
}

/*

Go 语言 中 new 和 make 是两个内置 函数，主要用来创建并分配内存。
Golang 中的 new 与 make 的区别是 new 只分配内存，而 make 只能用于 slice、map 和 channel 的初始化。

*new和make主要区别
	make 只能用来分配及初始化类型为 slice、map、chan 的数据，而 new 可以分配任意类型的数据。

	new 分配返回的是指针，即类型 *Type。make 返回引用，即 Type。

	new 分配的空间被清零。make 分配空间后，会进行初始化。
*
*/

func main() {

	s := new(Show) //new 创建Show对象，并只分配内存，返回 *Show 类型指针

	s.Param = make(map[string]any) //缺少 make分配内存空间，和初始化，直接对s.Param操作，编译会报错 "assignment to entry in nil map"

	s.Param["RMB"] = 1000

	fmt.Println(s.Param)
}

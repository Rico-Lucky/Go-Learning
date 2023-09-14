package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func main() {
	// go 的内置切片语法截取字符串 str[:]
	// 按字节截取，在处理 ASCII 单字节字符串截取，没有什么比这更完美的方案了

	a := "golang"
	fmt.Println(a[0:1]) //输出go
	fmt.Println(a[1:2]) //输出go

	// 中文往往占多个字节，在 utf8 编码中是3个字节
	//这是因为在Golang中string类型的底层是通过byte数组实现的,在unicode编码中,中文字符占两个字节,而在utf-8编码中,中文字符占三个字节而Golang的默认编码正是utf-8
	b := "go语言"
	fmt.Println(b[0:4]) //乱码
	fmt.Println(b[0:5]) //go语
	fmt.Println(b[2:])  //语言

	//类型转换 []rune
	rs := []rune(b)
	fmt.Println(string(rs[0:4]))

	tmp := "Go语言是Google开发的一种静态强类型、编译型、并发型，并具有垃圾回收功能的编程语言。"
	fmt.Println(tmp[:3])
	for i := range tmp {
		//range 是按字符迭代的，并非字节。下标从0开始
		if i == 20 {
			fmt.Println(tmp[:20])
			break
		}
	}

	name := "rico黄"
	fmt.Println(len(name)) //7

	//如果想要获得真实的字符串长度而不是其所占用字节数,有两种方法实现：
	// 方法一:使用unicode/utf-8包中的RuneCountInString方法
	fmt.Println(utf8.RuneCountInString(name)) //5
	// 方法二:将字符串转换为rune类型的数组再计算长度
	fmt.Println(len([]rune(name))) //5

	s := "SmaLLming张"
	//strings.Index(s, "m")查找字符“m”在字符串s中第一次出现的位置
	fmt.Println(strings.Index(s, "m")) //输出1
	//strings.LastIndex(s, "m")查找字符"m"在字符串s中最后一次出现的位置
	fmt.Println(strings.LastIndex(s, "m")) //输出5
	//strings.HasPrefix(s, "small")判断字符串s是否以指定字符串“small”开头
	fmt.Println(strings.HasPrefix(s, "small")) //输出true
	//strings.HasSuffix(s, "张")判断字符串s是否以指定字符串“张结尾”
	fmt.Println(strings.HasSuffix(s, "张")) //输出true
	//strings.Contains(s, "110")判断字符串s是否包含指定字符串“110”
	fmt.Println(strings.Contains(s, "110")) //输出false
	//strings.ToLower(s)将字符串s全部变小写
	fmt.Println(strings.ToLower(s)) //输出smallming张
	//strings.ToUpper(s)将字符串s全变大写
	fmt.Println(strings.ToUpper(s)) //输出SMALLMING张
	//strings.Replace(s, "m", "X",1)将字符串s中n的字符"m"替换成"X",当n小于0时表示全部替换
	fmt.Println(strings.Replace(s, "m", "X", -1)) //输出SXaLLXing张
	//strings.Repeat(s, 2)把字符串s复制count遍
	fmt.Println(strings.Repeat(s, 2)) //输出SmaLLming张SmaLLming张
	//strings.Trim(s, "S")去掉字符串前后指定的字符(前后只要有不管有几个就都去掉
	fmt.Println(strings.Trim(s, "S"))
	//当去掉空格时可以用strings.TrimSpace(s)代替
	fmt.Println(strings.TrimSpace(s))
	//strings.Split(s, "m")将s从指定字符"m"处切开,切片不再包括"m"
	fmt.Println(strings.Split(s, "m"))        //[S aLL ing张]
	fmt.Printf("%T\n", strings.Split(s, "m")) //[]string字符串切片类型
	//strings.Join(x, "")用指定分隔符将切片内容合并成字符串
	x := []string{"a", "b", "c"}
	fmt.Printf(strings.Join(x, "")) //输出abc

	x1 := []int{2, 3, 5, 7, 11}
	fmt.Println(x1[1:3])

}

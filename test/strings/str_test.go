package main

import (
	"fmt"
	"strings"
	"testing"
	"unicode"
)

//1.判断字符串中字符是否全都不同
//请实现⼀个算法，确定⼀个字符串的所有字符【是否全都不同】。这⾥我们要求【不允
// 许使⽤额外的存储结构】。 给定⼀个string，请返回⼀个bool值,true代表所有字符全都
// 不同，false代表存在相同的字符。 保证字符串中的字符为【ASCII字符】。字符串的⻓
// 度⼩于等于【3000】。

//解题思路

// 这⾥有⼏个重点，，第⼀个是 ASCII字符 ， ASCII字符 字符⼀共有256个，其中128个是常
// ⽤字符，可以在键盘上输⼊。128之后的是键盘上⽆法找到的。
// 然后是全部不同，也就是字符串中的字符没有重复的，再次，不准使⽤额外的储存结
// 构，且字符串⼩于等于3000。
// 如果允许其他额外储存结构，这个题⽬很好做。如果不允许的话，可以使⽤golang内置
// 的⽅式实现。

func IsUniqueSting1(str string) bool {

	//方法一：使用 strings.Count判断字符数量大于1时，代表有重复的字符
	if strings.Count(str, "") > 3000 {
		return false
	}

	for _, v := range str {

		if v > 127 {
			return false
		}

		if strings.Count(str, string(v)) > 1 {
			return false
		}

	}

	return true

}

func IsUniqueSting2(str string) bool {

	//方法2 通过 strings.Index 和 strings.LastIndex 函数判断：
	if strings.Count(str, "") > 3000 {
		return false
	}

	for k, v := range str {

		if v > 127 {
			return false
		}
		fmt.Println(strings.Index(str, string(v)))
		fmt.Println(strings.LastIndex(str, string(v)))
		fmt.Println(k)
		if strings.Index(str, string(v)) != k {

			return false
		}

	}

	return true
}

func TestIsUniqueSting1(t *testing.T) {

	fmt.Println(IsUniqueSting1("123saasdAd"))
	fmt.Println(IsUniqueSting1("12dA"))

}

func TestIsUniqueSting2(t *testing.T) {

	fmt.Println(IsUniqueSting2("122"))
	fmt.Println(IsUniqueSting2("12"))
}

/***
**

2.翻转字符串：
请实现⼀个算法，在不使⽤【额外数据结构和储存空间】的情况下，翻转⼀个给定的字
符串(可以使⽤单个过程变量)。
给定⼀个string，请返回⼀个string，为翻转后的字符串。保证字符串的⻓度⼩于等于
5000。
解题思路
翻转字符串其实是将⼀个字符串以中间字符为轴，前后翻转，即将str[len]赋值给str[0],
将str[0] 赋值 str[len]。

12 3 45
54 3 21

*/

func ReverString(str string) (string, bool) {

	strTem := []rune(str)
	l := len(strTem)
	if l > 5000 {
		return "", false
	}

	for i := 0; i < l/2; i++ {
		strTem[i], strTem[l-1-i] = strTem[l-1-i], strTem[i]
	}

	return string(strTem), true
}

func TestReverString(t *testing.T) {
	fmt.Println(ReverString("123456"))
}

// ####################
/*
3.判断两个给定的字符串排序后是否⼀致

问题描述
给定两个字符串，请编写程序，确定其中⼀个字符串的字符重新排列后，能否变成另⼀
个字符串。 这⾥规定【⼤⼩写为不同字符】，且考虑字符串重点空格。给定⼀个string
s1和⼀个string s2，请返回⼀个bool，代表两串是否重新排列后可相同。 保证两串的
⻓度都⼩于等于5000。
解题思路
⾸先要保证字符串⻓度⼩于5000。之后只需要⼀次循环遍历s1中的字符在s2是否都存
在即可。

*/

func IsRegroup(s1, s2 string) bool {

	s1R := len([]rune(s1))
	s2R := len([]rune(s2))

	if s1R > 5000 || s2R > 5000 || s1R != s2R {
		return false
	}

	for _, v := range s2 {
		if strings.Count(s1, string(v)) != strings.Count(s2, string(v)) {
			return false
		}
	}

	return true
}

func TestIsRegroup(t *testing.T) {
	fmt.Println(IsRegroup("abc", "bca"))
	fmt.Println(IsRegroup("abc", "bcd"))
	fmt.Println(IsRegroup("abc", "abcd"))
}

/*

4.字符串替换问题
问题描述
请编写⼀个⽅法，将字符串中的空格全部替换为“%20”。 假定该字符串有⾜够的空间存
放新增的字符，并且知道字符串的真实⻓度(⼩于等于1000)，同时保证字符串由【⼤⼩
写的英⽂字⺟组成】。 给定⼀个string为原始的串，返回替换后的string。


解题思路
两个问题，第⼀个是只能是英⽂字⺟，第⼆个是替换空格。


*/

func ReplaceBlank(s string) (string, bool) {

	if len([]rune(s)) > 1000 {
		return s, false
	}

	//只能是大小写英文字母组成
	for _, v := range s {
		if string(v) != " " && unicode.IsLetter(v) == false {
			return s, false
		}
	}

	return strings.Replace(s, " ", "%20", -1), true
	//  return strings.ReplaceAll(s, " ", "%20"), true
}

func TestReplaceBlank(t *testing.T) {

	fmt.Println(ReplaceBlank(" ASweeew "))
	fmt.Println(ReplaceBlank(" 1Sweeew "))

	// const (
	// 	A = iota
	// 	B
	// 	C
	// )

	// fmt.Println(A)
	// fmt.Println(B)
	// fmt.Println(C)

}

package main

import (
	"sync"
	"testing"
)

func TestSyncMap(t *testing.T) {

	var sMap sync.Map            //创建sync.Map实例
	sMap.Store("key1", "value1") //写
	sMap.Store("key2", "value2") //写
	sMap.Store("key3", "value3") //写
	sMap.Store("key4", "value4") //写
	// sMap.Delete("key1")          //删除

	// p, load := sMap.Swap("key1", "value1111") //key存在则交换新值，并返回旧值
	// if !load {
	// 	t.Error("key1 not exist")
	// 	return
	// }
	// tmpP := p.(string)
	// t.Errorf("tmpP:%+v", tmpP)

	// v, ok := sMap.Load("key1") //读
	// if !ok {
	// 	t.Error("key1 not exist")
	// 	return
	// }
	// str,_ := v.(string)  //断言根据类型取值
	// t.Errorf("v:%+v", str)

	//Range遍历 sync.Map 中每个kv，如果过滤器函数返回fasle,则结束遍历
	sMap.Range(func(key, value any) bool {
		t.Errorf("k:%v,v:%+v", key, value)
		return key != "key1"
	})

}

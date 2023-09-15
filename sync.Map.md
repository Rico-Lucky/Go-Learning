# 前言

- go version go1.20.4

# 1. sync.Map 是并发安全的map

- golang 中，**map不是安全的结构，并发读写会引发严重的错误**。

- sync标准包下的sync.Map能解决map并发读写的问题。



## 方法使用示例

```go
	var sMap sync.Map            //创建sync.Map实例
	sMap.Store("key1", "value1") //写
	sMap.Store("key2", "value2") //写
	sMap.Store("key3", "value3") //写
	sMap.Store("key4", "value4") //写
	sMap.Delete("key1")          //删除

	p, load := sMap.Swap("key1", "value1111") //key存在则交换新值，并返回旧值
	if !load {
		t.Error("key1 not exist")
		return
	}
	tmpP := p.(string)
	t.Errorf("tmpP:%+v", tmpP)

	v, ok := sMap.Load("key1") //读
	if !ok {
		t.Error("key1 not exist")
		return
	}
	str,_ := v.(string)  //断言根据类型取值
	t.Errorf("v:%+v", str)
	//Range遍历 sync.Map 中每个kv，如果过滤器函数返回fasle,则结束遍历
	sMap.Range(func(key, value any) bool {
		t.Errorf("k:%v,v:%+v", key, value)
		return key != "key1"
	})
```

# 2. 核心数据结构

## 2.1 sync.Map

sync.Map 主类中包含以下核心字段：

- • read：无锁化的只读 map，实际类型为 readOnly，2.3 小节会进一步介绍；
- • dirty：加锁处理的读写 map；
- • misses：记录访问 read 的失效次数，累计达到阈值时，会进行 read map/dirty map 的更新轮换；
- • mu：一把互斥锁，实现 dirty 和 misses 的并发管理.

可见，sync.Map 的特点是冗余了两份 map：read map 和 dirty map，后续的所介绍的交互流程也和这两个 map 息息相关，基本可以归结为两条主线：

主线一：首先基于无锁操作访问 read map；倘若 read map 不存在该 key，则加锁并使用 dirty map 兜底；

主线二：read map 和 dirty map 之间会交替轮换更新.

```go
type Map struct {
	mu Mutex 						//并发互斥锁，保护可读可写的map
	read atomic.Pointer[readOnly]   //无锁化的只读map
	dirty map[any]*entry 			//存在锁的读写map
	misses int 						//只读map miss次数
}
```

```go
type readOnly struct {
    m       map[any]*entry    //any 指interface{} 任何类型 无锁化的只读map
    amended bool // true if the dirty map contains some key not in m. 标识只读map是否缺失数据
}
```



## 2.2 entry对应的几种状态

kv 对中的 value，统一采用 unsafe.Pointer 的形式进行存储，通过 entry.p 的指针进行链接.

entry.p 的指向分为三种情况：

I 存活态：正常指向元素；

II 软删除态：指向 nil；

III 硬删除态：指向固定的全局变量 expunged.

- 存活态很好理解，即 key-entry 对仍未删除；

- nil 态表示软删除，read map 和 dirty map 底层的 map 结构仍存在 key-entry 对，但在逻辑上该 key-entry 对已经被删除，因此无法被用户查询到；

- expunged 态表示硬删除，dirty map 中已不存在该 key-entry 对.

  

```go
type entry struct {
	p atomic.Pointer[any]
}

```

```go
// A Pointer is an atomic pointer of type *T. The zero value is a nil *T.
type Pointer[T any] struct {
	// Mention *T in a field to disallow conversion between Pointer types.
	// See go.dev/issue/56603 for more details.
	// Use *T, not T, to avoid spurious recursive type definition errors.
	_ [0]*T

	_ noCopy
	v unsafe.Pointer
}
```

```go
var expunged = new(any)
```



##  2.3 readOnly

```go
type readOnly struct {
	m       map[any]*entry
	amended bool // true if the dirty map contains some key not in m.
}
```

sync.Map 中的只读 map：read 内部包含两个成员属性：

- m：真正意义上的 read map，实现从 key 到 entry 的映射；
- amended：标识 read map 中的 key-entry 对是否存在缺失，需要通过 dirty map 兜底.



# 3. sync.Map读流程

## 3.1 sync.Map.Load()

```go
func (m *Map) Load(key any) (value any, ok bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		// Avoid reporting a spurious miss if m.dirty got promoted while we were
		// blocked on m.mu. (If further loads of the same key will not miss, it's
		// not worth copying the dirty map for this key.)
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		return nil, false
	}
	return e.load()
}
```

-  查看 read map 中是否存在 key-entry 对，若存在，则直接读取 entry 返回；
- 倘若第一轮 read map 查询 miss，且 read map 不全，则需要加锁 double check；
- 第二轮 read map 查询仍 miss（加锁后），且 read map 不全，则查询 dirty map 兜底；
- 查询操作涉及到与 dirty map 的交互，misses 加一；
- 解锁，返回查得的结果.



## 3.2 entry.load()

```go
func (e *entry) load() (value any, ok bool) {
	p := e.p.Load()
	if p == nil || p == expunged {  //代表key-entry已被删除
		return nil, false
	}
	return *p, true
}
```

- sync.Map 中，kv 对的 value 是基于 entry 指针封装的形式；
-  从 map 取得 entry 后，最终需要调用 entry.load 方法读取指针指向的内容；
-  倘若 entry 的指针状态为 nil 或者 expunged，说明 key-entry 对已被删除，则返回 nil；
- 倘若 entry 未被删除，则读取指针内容，并且转为 any 的形式进行返回.





## 3.3 sync.Map.missLocked()

- 在读流程中，倘若未命中read map，且由于read map内容存在缺失需要和dirty map交互时，会走进missLocked流程。
- 在missLocked流程中
  - 首先计数器misses累加1；
  - 倘若misses次数小于dirty map 中存在的key-entry 对数量，直接返回即可；
  - miss次数大于等于dirty map 中存在的key-entry数量，则使用dirty map 覆盖read map ，置amended 为false；
  - 新的dirtymap置为nil；
  - missses 计数器清0。



```go
func (m *Map) missLocked() {
	m.misses++  				//计数器累加1
	if m.misses < len(m.dirty) {
		return   				//计数器小于dirty总数据量
	}
	m.read.Store(&readOnly{m: m.dirty}) 
	m.dirty = nil  
	m.misses = 0    
}
```



# 4. sync.Map写流程

## 4.1 sync.Map.Store()

- 调用 sync.Map.Swap方法

```go
// Store sets the value for a key.
func (m *Map) Store(key, value any) {
	_, _ = m.Swap(key, value)
}
```



## 4.2 sync.Map.Swap()

- 到read map中查找key-entry ，存在，则使用trySwap方法尝试更新
- 不存在read map 中，则加锁，double check 是否在只读read map 中，在则直接赋值
- 不在，则到 dirty map中

```go
func (m *Map) Swap(key, value any) (previous any, loaded bool) {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if v, ok := e.trySwap(&value); ok {
			if v == nil {
				return nil, false
			}
			return *v, true
		}
	}

	m.mu.Lock()  //加锁
	read = m.loadReadOnly()  //double check 是否在只读read map 中
	if e, ok := read.m[key]; ok {  //在 read map中
		if e.unexpungeLocked() {
			// The entry was previously expunged, which implies that there is a
			// non-nil dirty map and this entry is not in it.
			m.dirty[key] = e
		}
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else if e, ok := m.dirty[key]; ok {  //在dirty map 中
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else {
		if !read.amended { //新的 kv对插入dirty map
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			m.dirtyLocked()
			m.read.Store(&readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
	}
	m.mu.Unlock()
	return previous, loaded
}
```

（1）倘若 read map 存在拟写入的 key，且 entry 不为 expunged 状态，说明这次操作属于更新而非插入，直接基于 CAS 操作进行 entry 值的更新，并直接返回（存活态或者软删除，直接覆盖更新）；

（2）倘若未命中（1）的分支，则需要加锁 double check；

（3）倘若第二轮检查中发现 read map 或者 dirty map 中存在 key-entry 对，则直接将 entry 更新为新值即可（存活态或者软删除，直接覆盖更新）；

（4）在第（3）步中，如果发现 read map 中该 key-entry 为 expunged 态，需要在 dirty map 先补齐 key-entry 对，再更新 entry 值（从硬删除中恢复，然后覆盖更新）；

（5）倘若 read map 和 dirty map 均不存在，则在 dirty map 中插入新 key-entry 对，并且保证 read map 的 amended flag 为 true.（插入）

（6）第（5）步的分支中，倘若发现 dirty map 未初始化，需要前置执行 dirtyLocked 流程；

（7）解锁返回.  

## 4.3 entry.trySwap()

```go
func (e *entry) trySwap(i *any) (*any, bool) {
	for {
		p := e.p.Load()
		if p == expunged {
			return nil, false
		}
		if e.p.CompareAndSwap(p, i) {
			return p, true
		}
	}
}
```

- 在写流程中，倘若发现 read map 中已存在对应的 key-entry 对，则会对调用 trySwap方法尝试进行更新；
- 倘若 entry 为 expunged 态，说明已被硬删除，dirty 中缺失该项数据，因此 trySwap执行失败，回归主干流程；
- 倘若 entry 非 expunged 态，则直接执行 CAS 操作完成值的更新即可.





## 4.4 entry.unexpungeLocked()

```go
func (m *Map) Swap(key, value any) (previous any, loaded bool) {
//...
	m.mu.Lock()  //加锁
	read = m.loadReadOnly()  //double check 是否在只读read map 中
	if e, ok := read.m[key]; ok {  //在 read map中
		if e.unexpungeLocked() {
			// The entry was previously expunged, which implies that there is a
			// non-nil dirty map and this entry is not in it.
			m.dirty[key] = e
		}
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} //...
	m.mu.Unlock()
	return previous, loaded
}
func (e *entry) unexpungeLocked() (wasExpunged bool) {
	return e.p.CompareAndSwap(expunged, nil)
}
```

- 在写流程加锁 double check 的过程中，倘若发现 read map 中存在对应的 key-entry 对，会执行该方法；
- 倘若 key-entry 为硬删除 expunged 态，该方法会基于 CAS 操作将其更新为软删除 nil 态，然后进一步在 dirty map 中补齐该 key-entry 对，实现从硬删除到软删除的恢复.



## 4.5 entry.swapLocked()

```go
func (e *entry) swapLocked(i *any) *any {
	return e.p.Swap(i)
}
```

写流程中，倘若 read map 或者 dirty map 存在对应 key-entry，最终会通过原子操作，将新值的指针存储到 entry.p 当中.

## 4.6 sync.Map.dirtyLocked()

- 与missLocked 对偶流程
- 写多读少的场景会容易触发dirtyLocked

```go
func (m *Map) dirtyLocked() {
	if m.dirty != nil {  //代表ditry map 和 read map 拥有一样的数据，则不需要操作
		return
	}

    // 如果dirty map 为空，代表发生了missLocked,
    // 则需要把 read map 数据拷贝回到 ditry map,
    //重要的是回收非删除态
    //软删除变为硬删除
    //硬删除就不会拷贝到dirty map中.
	read := m.loadReadOnly()
	m.dirty = make(map[any]*entry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}

//....
func (e *entry) tryExpungeLocked() (isExpunged bool) {
	p := e.p.Load()
	for p == nil {
		if e.p.CompareAndSwap(nil, expunged) {
			return true
		}
		p = e.p.Load()
	}
	return p == expunged
}
```

- 在写流程中，倘若需要将 key-entry 插入到兜底的 dirty map 中，并且此时 dirty map 为空（从未写入过数据或者刚发生过 missLocked），会进入 dirtyLocked 流程；
- 此时会遍历一轮 read map ，将未删除的 key-entry 对拷贝到 dirty map 当中；
- •在遍历时，还会将 read map 中软删除 nil 态的 entry 更新为硬删除 expunged 态，因为在此流程中，不会将其拷贝到 dirty map.



# 5. 删除

## 5.1 sync.Map.Delete()

```go
// Delete deletes the value for a key.
func (m *Map) Delete(key any) {
	m.LoadAndDelete(key)
}
```



## 5.2 sync.Map.LoadAndDelete()

```go
func (m *Map) LoadAndDelete(key any) (value any, loaded bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			delete(m.dirty, key) //物理删除
			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			m.missLocked()  //增加misses次数，次数达到了dirty map 的长度容量，则触发拷贝到read map
		}
		m.mu.Unlock()
	}
	if ok {
		return e.delete()
	}
	return nil, false
}
```

（1）倘若 read map 中存在 key，则直接基于 cas 操作将其删除；

（2）倘若read map 不存在 key，且 read map 有缺失（amended flag 为 true），则加锁 dou check；

（3）倘若加锁 double check 时，read map 仍不存在 key 且 read map 有缺失，则从 dirty map 中取元素，并且将 key-entry 对从 dirty map 中物理删除；

（4）走入步骤（3），删操作需要和 dirty map 交互，需要走进 3.3 小节介绍的 missLocked 流程；

（5）解锁；

（6）倘若从 read map 或 dirty map 中获取到了 key 对应的 entry，则走入 entry.delete() 方法逻辑删除 entry；

（7）倘若 read map 和 dirty map 中均不存在 key，返回 false 标识删除失败.  

## 5.3 entry.delete()

```go
func (e *entry) delete() (value any, ok bool) {
	for {
		p := e.p.Load()
		if p == nil || p == expunged {  //已被软删除 或 硬删除
			return nil, false
		}
		if e.p.CompareAndSwap(p, nil) {
			return *p, true
		}
	}
}
```

- 该方法时entry的逻辑删除方法
- 倘若entry此前已被删除，则直接返回false 标识删除失败
- 倘若enrty 当前仍存在，则通过CAS讲entry.p指向nil ,标识其已进入软删除状态。







# 6. 遍历

- 用户传进熔断器函数，熔断器函数入参为kv对，函数返回false 则结束遍历
- 可能会显式地触发dirty map 全量拷贝到 read map，置dirty map 为nil
  - 下一次插入操作，会触发 dirtyLocked（）内流程，read map全量拷贝到dirty map 

```go
func (m *Map) Range(f func(key, value any) bool) {
	// We need to be able to iterate over all of the keys that were already
	// present at the start of the call to Range.
	// If read.amended is false, then read.m satisfies that property without
	// requiring us to hold m.mu for a long time.
	read := m.loadReadOnly()
    if read.amended {   //执行与missLocked()相同的动作
		// m.dirty contains keys not in read.m. Fortunately, Range is already O(N)
		// (assuming the caller does not break out early), so a call to Range
		// amortizes an entire copy of the map: we can promote the dirty copy
		// immediately!
		m.mu.Lock()
		read = m.loadReadOnly()
		if read.amended {
			read = readOnly{m: m.dirty}   // 触发dirty map 全量拷贝到 read map
			m.read.Store(&read)
			m.dirty = nil                
			m.misses = 0 
		}
		m.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}
```



# 7. 软删除(nil)硬删除(expunged)

## 7.1 entry 的expunged

首先需要明确，无论是软删除（nil）还是硬删除 （expunged），都代表在逻辑上key-entry对已从sync.Map中删除，nil 和 expunged的区别在于：

- 软删除态（nil）：read map 和 dirty map 在物理上仍保有该key-entry对，因此倘若此时需要对该entry执行写操作，可以直接CAS操作。
- 硬删除态（expunged）：dirty map 中已经没有该key -entry对，倘若执行写操作，必须加锁（diry map 必须含有全量key-entry对数据）

设计 nil 和 expunged 两种状态的原因，就是为了优化在 dirtyLocked前，针对同一个key先删后写的场景，通过expunged态额外标识出dirty map 中是否仍具有指向entry的能力，这样能实现对一部分nil态 key-entry 对的解放，能够基于CAS完成这部分内容写入操作和无需加锁。



## 7.2 read map 和 dirty map 的数据流转

sync.Map 由两个map构成：

- read map ：访问时全程无锁。
- dirty map : 兜底的读写map，访问时需要加锁

之所以这样处理，是希望能够根据读、删、更新、写操作频次的探测，来实时地调整操作方式，希望在读、更新、删频次更高时，更好地采用CAS的方式无锁化地完成操作；在写操作频次更高时，则直接采用加锁操作完成。

因此，sync.Map 本质上采用了一种以空间换时间+动态调整策略的设计思路。

## 7.2.1 两个 map

-  总体思想，希望能多用 read map，少用 dirty map，因为操作前者无锁，后者需要加锁；
- 除了 expunged 态的 entry 之外，read map 的内容为 dirty map 的子集；

## 7.2.2 dirty map -> read map

记录读/删流程中，通过 misses 记录访问 read map miss 由 dirty 兜底处理的次数，当 miss 次数达到阈值，则进入 missLocked 流程，进行新老 read/dirty 替换流程；此时将老 dirty 作为新 read，新 dirty map 则暂时为空，直到 dirtyLocked 流程完成对 dirty 的初始化；

## 7.2.3 read map -> dirty map

- 发生 dirtyLocked 的前置条件：I dirty 暂时为空（此前没有写操作或者近期进行过 missLocked 流程）；II 接下来一次写操作访问 read 时 miss，需要由 dirty 兜底；
- 在 dirtyLocked 流程中，需要对 read 内的元素进行状态更新，因此需要遍历，是一个线性时间复杂度的过程，可能存在性能抖动；
- dirtyLocked 遍历中，会将 read 中未被删除的元素（非 nil 非 expunged）拷贝到 dirty 中；会将 read 中所有此前被删的元素统一置为 expunged 态.

#  7.3 sync.Map适用场景 和 注意问题

-  sync.Map 适用于读多、更新多、删多、写少的场景。
- 倘若写操作过多，sync.Map基本等价于互斥锁+map
- sync.Map可能存在性能抖动问题，主要发生在读/删流程miss只读 map 次数过多时，触发missLocked流程，下一次插入操作的过程当中触发dirtyLocked流程。
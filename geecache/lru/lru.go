package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // 允许使用的最大内存
	nBytes    int64                         // 当前使用内存
	cache     map[string]*list.Element      // 字典
	ll        *list.List                    // 双向链表
	OnEvicted func(key string, value Value) // 记录被移除时的回调函数，可以为 nil
}

type entry struct { // 双向链表存储的数据结构
	key   string
	value Value // 值为实现了Value接口的任意类型
}

type Value interface {
	Len() int // 返回值所占内存大小
}

func (c *Cache) Len() int {
	return c.ll.Len()
}

// 为了方便实例化，实现 New() 函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		cache:     make(map[string]*list.Element),
		ll:        list.New(),
		OnEvicted: onEvicted,
	}
}

// 查找：从字典中找到双向链表对应的节点，然后将该节点移动到队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 存在节点，将节点移动到队尾
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry) // !此处获取到的是list源码包中Element的Value
		return kv.value, true
	}
	// 没有该节点
	return
}

// 删除：缓存淘汰，移除最近最少访问的节点（队首）
func (c *Cache) RemoveOldest() {
	// 取到队首元素
	ele := c.ll.Front()
	if ele != nil {
		// 删除队首元素
		c.ll.Remove(ele)
		// 从字典中 c.cache 删除该节点的映射关系
		kv := ele.Value.(*entry) // 通过双向链表的值为entry, 这里使用了类型断言(因为Value实现了任意接口类型，可以存储任意值所以可以使用断言)
		delete(c.cache, kv.key)
		// 更新当前占用内存
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 新增/修改：键存在则更新键，并将其移动到队尾
// 键不存在则向队尾添加新节点，并在字典中添加 key 和 节点的映射关系
// 更新占用内存，如果超出最大内存，则删除队首元素
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len()) // 加上新值的内存，剪掉旧值的内存
		kv.value = value
	} else {
		ele := c.ll.PushBack(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

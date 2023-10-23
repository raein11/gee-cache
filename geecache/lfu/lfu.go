package lfu

import "container/list"

type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	onEvicted func(key string, value interface{})
}

type entry struct {
	key   string
	value interface{}
	count int // 记录当前 key 的缓存被访问次数
	index int // 堆索引
}

func (c *Cache) Len() int {
	return c.ll.Len()
}

func New(maxBytes int64, onEvicted func(string, interface{})) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element, 0),
		onEvicted: onEvicted,
	}
}

// 查找：根据 key 获取缓存值，每次查找将使用次数 +1

package singleflight

import "sync"

// 正在进行，或者已经结束的请求，使用 sync.WaitGroup 避免重入
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// singleflight 主数据结构，管理不同的请求（call）
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// DO 方法保证 key 相同时，访问远程节点只发起一次请求
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 1.对 g.m 进行操作前上锁
	g.mu.Lock()
	// 2.检查 g.m 是否为 nil
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 3.尝试获取对应的 call 对象
	if c, ok := g.m[key]; ok {
		// 4.存在则释放锁并等待计算完成，返回计算结果
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	// 5.不存在则创建一个新的 call 对象，将其与 key 关联，释放读写锁
	c := new(call) // new 返回的是指针
	c.wg.Add(1)    // 发起请求前加锁
	g.m[key] = c
	g.mu.Unlock()
	// 6.释放锁后，调用 fn，并将结果保存到 c.val, c.err
	c.val, c.err = fn() // 发起请求
	// 7.计算完成后调用 c.wg.Done() 通知等待中的 call
	c.wg.Done() // 请求结束
	// 8.上锁，删除计算完成的 call 对象，释放锁
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	// 9.返回计算结果 val, err
	return c.val, c.err
}

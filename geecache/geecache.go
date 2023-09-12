package geecache

import "sync"

// Getter 接口
type Getter interface {
	// Get 回调函数
	Get(key string) ([]byte, error)
}

// 定义函数类型 GetterFunc, 并实现 Getter 接口的 Get 方法
type GetterFunc func(key string) ([]byte, error)

// 函数类型实现接口（接口函数）
// 方便使用者在调用时既能够传入函数作为参数， 也能够传入实现了该接口的结构体作为参数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string // 唯一的名称
	getter    Getter // 缓存未命中时的回调
	mainCache cache  // 并发缓存
}

// 全局变量
var (
	mu     sync.Mutex
	groups = make(map[string]*Group)
)

// 构造函数，用于实例化 Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	// 1.判断是否传入回调函数
	if getter == nil {
		panic("nil Getter")
	}
	// 考虑到并发问题，先上锁
	mu.Lock()
	defer mu.Unlock() // 延迟释放：在 return 之后，函数返回结果之前
	// 构造 Group
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

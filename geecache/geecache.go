package geecache

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

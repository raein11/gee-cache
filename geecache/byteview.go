package geecache

// 只读数据结构 ByteView 用来表示缓存值
type ByteView struct {
	b []byte // 使用 byte 类型可以支持任意数据类型的存储，如字符串、图片等
}

// Len 方法 获取缓存的大小
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 方法返回一个拷贝，防止缓存值被外部程序修改。
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// cloneBytes 函数返回缓存的拷贝，防止外部程序修改，此方法不对外提供，内部自己调
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// String 方法返回字符串类型的缓存
func (v ByteView) String() string {
	return string(v.b)
}

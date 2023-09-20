package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 1.定义函数类型 Hash，采取依赖注入的方式，允许用于替换自定义的 Hash函数，默认为 crc32.ChecksumIEEE 算法
type Hash func(data []byte) uint32

// 2.定义一致性哈希算法的主数据结构 Map
type Map struct {
	hash     Hash           // Hash 函数 hash
	replicas int            // 虚拟节点倍数 replicas
	keys     []int          // 哈希环 keys
	hashMap  map[int]string // 虚拟节点与真实节点的映射 hashMap
}

// 3.实现 Map 构造函数 New
// 允许自定义虚拟节点倍数，Hash 函数
func New(replicas int, fn Hash) *Map {
	// 1.初始化 m
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	// 2.如果没有传入 Hash 函数，则使用默认的
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE // 源码中该函数接受 []byte 类型的参数进行 hash 计算
	}
	// 3.返回
	return m
}

// 4.实现 Add 方法
func (m *Map) Add(keys ...string) {
	// 1.遍历真实节点，添加虚拟节点
	for _, key := range keys {
		// 每个真实节点对应倍数个虚拟节点
		for i := 0; i < m.replicas; i++ {
			// 将节点进行 hash 计算并转换类型, 使用编号+真实节点名称作为虚拟节点
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 将虚拟节点加入到哈希环
			m.keys = append(m.keys, hash)
			// 添加虚拟节点和真实节点的映射
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

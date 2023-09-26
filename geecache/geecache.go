package geecache

import (
	"log"
	"sync"
)

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
	name      string     // 唯一的名称
	getter    Getter     // 缓存未命中时的回调
	mainCache cache      // 并发缓存
	peers     PeerPicker // 分布式节点
}

// 全局变量
var (
	mu     sync.RWMutex // 读写锁
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

// GetGroup: 根据 name 获取 Group
func GetGroup(name string) *Group {
	// 减小锁的粒度，这里使用只读锁，因为没有涉及到修改操作
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// group 的 Get 方法：实现返回缓存值（1）和返回缓存值（3）
func (g *Group) Get(key string) (ByteView, error) {
	// 1.去除传入 key 为空的情况
	if key == "" {
		return ByteView{}, nil
	}
	// 2.判断情况（1）
	if v, ok := g.mainCache.Get(key); ok {
		// 缓存命中
		log.Println("缓存命中")
		return v, nil
	}
	// 缓存未命中
	return g.load(key)
}

//	func (g *Group) load(key string) (ByteView, error) {
//		return g.getLocally(key)
//	}
//
// 修改后的 load 方法 使用 PickPeer 方法获取节点
func (g *Group) load(key string) (value ByteView, err error) {
	// 1.判断是否存在节点
	if g.peers != nil {
		// 2.使用 PickPeer 选择节点
		if peer, ok := g.peers.PickPeer(key); ok {
			// 3.尝试根据远程节点获取缓存值
			if value, err = g.getFromPeer(peer, key); err != nil {
				// 4.返回从远程获取的节点
				return value, err
			}
		}
	}
	// 5.是本机节点或从远程节点获取失败则调用 getLocally 方法
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用回调函数获取源数据
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// 调用缓存克隆方法，封装数据
	value := ByteView{cloneBytes(bytes)}
	g.populateGroup(key, value)
	return value, nil
}

// 将数据添加到缓存
func (g *Group) populateGroup(key string, value ByteView) {
	g.mainCache.Add(key, value)
}

// 将实现了 PeerPicker 的 HTTPPool 注入到 Group 中
func (g Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers 多次调用")
	}
	g.peers = peers
}

// 使用实现了 PeerGetter 接口的 httpGetter 访问远程节点，获取缓存值
func (g Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

package geecache

import pb "geecache/geecachepb"

// 1.PeerPicker 接口
type PeerPicker interface {
	// PickPeer() 方法根据传入的 key 选择相应的节点 PeerGetter
	PickPeer(key string) (peerGetter PeerGetter, ok bool)
}

// 2.PeerGetter 接口
type PeerGetter interface {
	// 修改前
	// Get() 方法根据 key 从对应的 group 中查找缓存（PeerGetter 对应 HTTP 客户端）
	//Get(group string, key string) ([]byte, error)

	// 修改后
	Get(in *pb.Request, out *pb.Response) error
}

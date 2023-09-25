package geecache

import (
	"fmt"
	"geecache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map    // 用于根据 key 选择节点
	httpGetters map[string]*httpGetter // 映射远程节点与对应的 httpGetter
	// 一个远程节点对应一个 httpGetter，因为 httpGetter 与远程节点的地址 baseURL 有关
}

// 初始化 HTTPPool
func NewHTTPPoll(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// 日志打印方法
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1.判断访问路径的前缀是否是 basePath
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		// 前缀不匹配
		panic("HTTPPoll 提供的路径不匹配：" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// 2.获取请求路径中去掉前缀的部分
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "错误请求", http.StatusBadRequest)
	}
	// 约定访问路径为 /<basepath>/<groupname>/<key>
	// 3.获取 groupname, key
	groupName := parts[0]
	key := parts[1]

	// 4.通过名称尝试获取 group
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "未获取到 group："+groupName, http.StatusNotFound)
		return
	}

	// 5.group 不为 nil, 尝试获取 key
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 6.到这里，获取到了缓存，设置响应头，返回类型为文件字节流
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// HTTP 客户端类 httpGetter
type httpGetter struct {
	baseURL string
}

// httpGetter 实现 PeerGetter接口
var _ PeerGetter = (*httpGetter)(nil)

// 在 httpGetter 类型的值上断言 PeerGetter 接口，断言成功则实现接口
// 可以在编译时检查 httpGetter 是否实现 PeerGetter 接口

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	// 1.拼接访问路径，为了安全使用 url.QueryEscape 对字符串进行转义
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u) // 获取请求响应
	// 2.请求是否异常
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	// 3.请求正常，判断响应状态码是否 OK
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务端返回：%v", res.StatusCode)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	// 4.状态码 OK，判断响应体是否为空
	if err != nil {
		return nil, fmt.Errorf("获取响应体：%v", err)
	}
	// 5.返回响应体
	return bytes, nil
}

func (p *HTTPPool) Set(peers ...string) {
	// 1.上锁
	p.mu.Lock()
	defer p.mu.Unlock()
	// 2.实例化一致性哈希算法
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	// 3.添加传入的节点
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

var _ PeerPicker = (*HTTPPool)(nil)

func (p *HTTPPool) PickPeer(key string) (peerGetter PeerGetter, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" || peer != p.self {
		p.Log("选择节点 %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

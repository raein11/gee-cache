package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache/"

type HTTPPool struct {
	self     string
	basePath string
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
	log.Println("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1.判断访问路径的前缀是否是 basePath
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		// 前缀不匹配
		panic("HTTPPoll 提供的路径不匹配：" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// 获取请求路径中去掉前缀的部分
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "错误请求", http.StatusBadRequest)
	}
	// 约定访问路径为 /<basepath>/<groupname>/<key>
	// 获取 groupname, key
	groupName := parts[0]
	key := parts[1]

	// 通过名称尝试获取 group
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "未获取到 group："+groupName, http.StatusNotFound)
		return
	}

	// group 不为 nil, 尝试获取 key
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 到这里，获取到了缓存
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

package main

import (
	"flag"
	"fmt"
	"geecache"
	"log"
	"net/http"
)

// 使用 map 模拟数据源 db
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 封装 createGroup 函数用于创建 Group
func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s 不存在", key)
		}))
}

// 实现缓存服务器 startCacheServer 函数
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	// 1.创建一个 HTTPPool
	peers := geecache.NewHTTPPool(addr)
	// 2.使用一致性哈希算法添加节点
	peers.Set(addrs...)
	// 3.注册节点到 Group
	gee.RegisterPeers(peers)
	log.Println("geecache 运行在", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers)) // http://xxx
}

// 实现 API 服务器 startAPIServer
func startAPIServer(apiAddr string, gee *geecache.Group) {
	// 1.实现 http.Handle 方法，处理后缀 /api 的请求
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// 1.取出请求路径中表示 key 的部分
			key := r.URL.Query().Get("key")
			// 2.查找 key 对应的缓存
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// 设置响应体的请求头
			w.Header().Set("Content-Type", "aplication/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server 运行在", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
	// 传入 nil 默认使用多路复用来处理请求
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache 服务器端口")
	flag.BoolVar(&api, "api", false, "启用 api 服务器？")
	flag.Parse()
	// 1.初始化参数
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	// 2.创建一个名为 scores 的 Group，若缓存为空， 回调函数会从 db 中获取数据并返回
	gee := createGroup()

	// 3.判断是否启动 api 服务器
	if api {
		go startAPIServer(apiAddr, gee) // api 服务器只有一个，缓存服务器有多个，这里需要使用协程
	}
	// 4.启动缓存服务器
	startCacheServer(addrMap[port], []string(addrs), gee) // 这里使用 []string(addrs) 会创建一个新的切片，底层数组不会共享
}

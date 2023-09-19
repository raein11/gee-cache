package main

import (
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

func main() {
	// 1.创建一个名为 scores 的 Group，若缓存为空， 回调函数会从 db 中获取数据并返回
	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s 不存在", key)
		}))
	// 2.初始化参数
	addr := "localhost:9999"
	peers := geecache.NewHTTPPoll(addr)
	log.Println("geecache is running at", addr)
	// 3.在 addr 启动 HTTP 服务
	log.Fatal(http.ListenAndServe(addr, peers))
}

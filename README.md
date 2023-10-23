# GeeCache
基于 GoLang 的简易分布式缓存系统

项目目录

```
gee-cache
│  go.mod
│  go.sum
│  LICENSE
│  main.go // 缓存服务器、API 服务器
│  README.md
│  run.sh // 服务运行脚本
│
└─geecache
    │  byteview.go // 只读数据结构
    │  cache.go // 缓存封装
    │  geecache.go // 主数据结构
    │  geecache_test.go
    │  go.mod
    │  go.sum
    │  http.go // 封装 HTTPPool
    │  peers.go // 抽象接口
    │
    ├─consistenthash // 一致性哈希
    │      consistenthash.go
    │      consistenthash_test.go
    │
    ├─geecachepb // protobuf
    │      geecachepb.pb.go
    │      geecachepb.proto
    │
    ├─lru // LRU 淘汰算法
    │      lru.go
    │      lru_test.go
    │
    ├─lfu // LFU 淘汰算法
    |      lfu.go
    |      lfu
    |
    └─singleflight // 防止缓存击穿
```


package geecache

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("回调失败！")
	}
}

// 模拟耗时的数据库
var db = map[string]string{
	"name":     "root",
	"password": "123455",
	"driver":   "mysql",
}

func TestGet(t *testing.T) {
	//  统计回调次数，测试缓存存在的情况下是否会调用回调函数
	loadCounts := make(map[string]int, len(db))
	// 构造 Group，构造完成后 每个 key 的 loadCount 的值为 1
	// 第一次测试失败 这里设置缓存大小为 8 测试不通过，因为后续加入缓存时空间不足
	gee := NewGroup("connection", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		fmt.Println("[slowDB] search key..", key)
		// 尝试获取缓存
		if v, ok := db[key]; ok {
			// 获取到 key 了, 调用了回调函数
			// 如果是第一次获取该值的 key 的话，则将值置为0
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			// 每次调用 +1操作
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return nil, fmt.Errorf("回调未获取到缓存")
	}))
	for k, v := range db {

		// 正常情况，走第一种情况，缓存命中，所以不会进入判断
		if view, err := gee.Get(k); err != nil || view.String() != v {
			//	如果有值的情况下没获取到值说明测试失败
			t.Fatalf("未能获取到值")
		}
		// 正常情况，每个缓存的调用回调函数的次数都为 1
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			// 走到这里说明有缓存还调用了回调函数
			t.Fatalf("缓存丢失")
		}
	}
	// 测试获取不存在的缓存，能否通过回调函数获取源数据
	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("这个值应该为空，但 %s 通过回调函数获取到了", view)
	}
}

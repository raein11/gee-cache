package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// 1.构造自定义的一致性哈希
	hash := New(3, func(key []byte) uint32 {
		// 测试需要知道每一个传入的 key 的哈希值，使用默认的 crc32.ChecksumIEEE 算法达不到，所以这里使用自定义的哈希函数
		// 自定义的哈希函数，传入字符串类型的数字，再将数字返回
		num, _ := strconv.Atoi(string(key))
		return uint32(num)
	})
	// 2.初始化三个真实节点 2/4/6，分别对应三组虚拟节点（编号+真实节点名称） 02/12/22、04/14/24、06/16/26。
	hash.Add("2", "4", "6")
	// 3.创建用例 2/11/23/27，对应的虚拟节点 02/12/24/02，对应真实节点 2/2/4/2。
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}
	// 4.添加真实节点 8，对应虚拟节点 08/18/28
	// 此时 "27" 对应的虚拟节点变为 28，真实节点变为 8。
	hash.Add("8")
	testCases["27"] = "8"
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("请求 %s, 应该产生 %s", k, v)
		}
	}
}

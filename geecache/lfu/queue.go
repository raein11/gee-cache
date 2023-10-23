package lfu

import "container/heap"

type Queue []*entry

func (q Queue) Len() int {
	return len(q)
}

func (q Queue) Less(i, j int) bool {
	return q[i].count < q[j].count
}

func (q Queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	// 注意将索引交换回来
	q[i].index = i
	q[j].index = j
}

func (q *Queue) Push(v interface{}) {
	n := q.Len()
	en := v.(*entry)
	en.index = n
	*q = append(*q, en) // 这里会重新分配内存，并拷贝数据
}

func (q *Queue) Pop() interface{} {
	oldEn := *q
	n := oldEn.Len()
	en := oldEn[n-1]
	oldEn[n-1] = nil // 将不再使用的对象置为nil，加快垃圾回收，避免内存泄漏
	*q = oldEn[:n-1] // 这里会重新分配内存，并拷贝数据
	return en
}

func (q *Queue) update(en *entry, val interface{}, count int) {
	// 更新缓存值和访问次数
	en.value = val
	en.count = count
	(*q)[en.index] = en
	// 重新排序
	heap.Fix(q, en.index)
}

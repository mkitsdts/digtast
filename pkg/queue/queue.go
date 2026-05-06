package queue

import "sync"

type Queue[T any] struct {
	items []T
	mux   *sync.Mutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		mux:   &sync.Mutex{},
		items: make([]T, 0),
	}
}

func (q *Queue[T]) Push(item T) {
	q.mux.Lock()
	defer q.mux.Unlock()
	q.items = append(q.items, item)
}

func (q *Queue[T]) Pop() (T, bool) {
	q.mux.Lock()
	defer q.mux.Unlock()
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}

	// 获取头部元素
	item := q.items[0]

	// 手动清零，帮助 GC 回收旧引用的内存
	var zero T
	q.items[0] = zero

	q.items = q.items[1:]
	return item, true
}

// 获取队首元素
func (q *Queue[T]) Front() (T, bool) {
	q.mux.Lock()
	defer q.mux.Unlock()
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}
	return q.items[0], true
}

// 检查队列是否为空
func (q *Queue[T]) Empty() bool {
	q.mux.Lock()
	defer q.mux.Unlock()
	return len(q.items) == 0
}

// 返回队列长度
func (q *Queue[T]) Size() int {
	q.mux.Lock()
	defer q.mux.Unlock()
	return len(q.items)
}

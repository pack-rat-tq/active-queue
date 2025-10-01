package main

import (
	"errors"
	"sync"
	"sync/atomic"
)

type ActiveQueue[T any] struct {
	head   *dataNode[T]
	tail   *dataNode[T]
	count  atomic.Int64
	out    chan T
	active bool
	l      sync.Mutex
	pool   sync.Pool
}

type dataNode[T any] struct {
	task T
	next *dataNode[T]
}

func NewActiveQueue[T any]() *ActiveQueue[T] {
	ret := new(ActiveQueue[T])
	ret.out = make(chan T)
	ret.pool = sync.Pool{
		New: func() interface{} {
			return &dataNode[T]{}
		},
	}
	return ret
}

func (q *ActiveQueue[T]) getNode() *dataNode[T] {
	return q.pool.Get().(*dataNode[T])
}

func (q *ActiveQueue[T]) putNode(n *dataNode[T]) {
	n.Reset()
	q.pool.Put(n)
}

func (n *dataNode[T]) Reset() {
	var def T
	n.task = def
	n.next = nil
}

func (q *ActiveQueue[T]) outWorker() {
	for {
		q.l.Lock()
		if q.head != nil {
			ret := q.head.task
			drop := q.head
			q.head = q.head.next
			q.putNode(drop)
			q.l.Unlock()
			q.out <- ret
			q.count.Add(-1)
		} else {
			q.active = false
			q.l.Unlock()
			return
		}
	}
}

func (q *ActiveQueue[T]) Push(task T) {
	q.l.Lock()
	defer q.l.Unlock()
	node := q.getNode()
	node.task = task
	q.count.Add(1)
	if q.head == nil {
		q.head = node
		q.tail = node
		if !q.active {
			q.active = true
			go q.outWorker()
		}
		return
	}
	q.tail.next = node
	q.tail = q.tail.next
}

func (q *ActiveQueue[T]) Pop() (T, error) {
	select {
	case ret, ok := <-q.out:
		if !ok {
			var def T
			return def, errors.New("not ready")
		}
		return ret, nil
	default:
	}
	var def T
	return def, errors.New("empty")
}

func (q *ActiveQueue[T]) Count() int64 {
	return q.count.Load()
}

func (q *ActiveQueue[T]) Receive() <-chan T {
	return q.out
}

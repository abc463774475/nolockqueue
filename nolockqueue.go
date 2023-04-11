package nolockqueue

import "sync/atomic"

type QueueData struct {
	next, prev *QueueData
	Value      interface{}
}

func New(value interface{}) *QueueData {
	r := new(QueueData)
	r.next = r
	r.prev = r
	r.Value = value
	return r
}

func (r *QueueData) Next() *QueueData {
	return r.next
}

func (r *QueueData) Prev() *QueueData {
	return r.prev
}

func (r *QueueData) Move(n int) *QueueData {
	if n < 0 {
		return r.Move(-n).Prev()
	}
	for ; n > 0; n-- {
		r = r.Next()
	}
	return r
}

func (r *QueueData) MovePrev(n int) *QueueData {
	return r.Move(-n)
}

func (r *QueueData) MoveNext(n int) *QueueData {
	return r.Move(n)
}

func (r *QueueData) Link(s *QueueData) *QueueData {
	n := r.Next()
	p := s.Prev()
	r.next = s
	s.prev = r
	n.prev = p
	p.next = n
	return r
}

func (r *QueueData) Unlink(num int) *QueueData {
	if num <= 0 {
		return r
	}
	p := r.Prev()
	n := r.Move(num)
	p.next = n
	n.prev = p
	return n
}

func (r *QueueData) Push(value interface{}) *QueueData {
	return r.Link(New(value))
}

func (r *QueueData) Pop() (*QueueData, interface{}) {
	next := r.Unlink(1)
	if next == nil {
		return nil, nil
	}
	return next, next.Value
}

func (r *QueueData) Remove() *QueueData {
	return r.Unlink(1)
}

func (r *QueueData) Do(f func(interface{})) {
	f(r.Value)
	next := r.Next()
	for next != nil && next != r {
		f(next.Value)
		next = next.Next()
	}
}

// 无锁队列
// 无锁队列是一种特殊的队列，它的特点是在多线程环境下，不需要加锁，也不会出现竞争。
type Queue struct {
	head, tail *QueueData

	// 用于标记队列是否为空
	// 0: 队列为空
	// 1: 队列不为空
	notEmpty int32

	// 用于标记队列是否已经关闭
	// 0: 队列未关闭
	// 1: 队列已经关闭
	closed int32

	// 用于标记队列是否已经销毁
	// 0: 队列未销毁
	// 1: 队列已经销毁
	destroyed int32

	// 用于标记队列是否已经初始化
	// 0: 队列未初始化
	// 1: 队列已经初始化
	initialized int32

	// 用于标记队列中的元素个数
	len int32

	// cas操作的锁
	lock int32
}

// 创建一个新的无锁队列
func NewQueue(v interface{}) *Queue {
	q := new(Queue)

	q.head = New(v)
	q.tail = q.head
	q.notEmpty = 1
	q.closed = 0
	q.destroyed = 0
	q.initialized = 1
	q.len = 1

	return q
}

// push操作
func (q *Queue) Push(v interface{}) {
	if q.destroyed == 1 {
		return
	}

	for {
		if q.closed == 1 {
			return
		}

		if q.lock == 0 && atomic.CompareAndSwapInt32(&q.lock, 0, 1) {
			break
		}
	}

	q.tail.Push(v)
	q.tail = q.tail.Next()
	atomic.AddInt32(&q.len, 1)
	q.notEmpty = 1

	q.lock = 0
}

// pop操作
func (q *Queue) Pop() interface{} {
	if q.destroyed == 1 {
		return nil
	}

	for {
		if q.closed == 1 {
			return nil
		}

		if q.lock == 0 && atomic.CompareAndSwapInt32(&q.lock, 0, 1) {
			break
		}
	}

	if q.len == 0 {
		q.lock = 0
		return nil
	}

	//v := q.head.Pop()
	//q.head = q.head.Next()
	t1, v := q.head.Pop()
	q.head = t1

	atomic.AddInt32(&q.len, -1)

	if q.len == 0 {
		q.notEmpty = 0
	}

	q.lock = 0
	return v
}

// 获取队列中的元素个数
func (q *Queue) Len() int {
	return int(q.len)
}

// 关闭队列
func (q *Queue) Close() {
	q.closed = 1
}

// 销毁队列
func (q *Queue) Destroy() {
	q.destroyed = 1
}

// 遍历队列中的元素
func (q *Queue) Do(f func(interface{})) {
	if q.destroyed == 1 {
		return
	}

	for {
		if q.closed == 1 {
			return
		}

		if q.lock == 0 && atomic.CompareAndSwapInt32(&q.lock, 0, 1) {
			break
		}
	}

	if q.len == 0 {
		q.lock = 0
		return
	}

	q.head.Do(f)

	q.lock = 0
}

// Value is the value stored in a Queue.
func (q *Queue) Value() interface{} {
	if q.destroyed == 1 {
		return nil
	}

	for {
		if q.closed == 1 {
			return nil
		}

		if q.lock == 0 && atomic.CompareAndSwapInt32(&q.lock, 0, 1) {
			break
		}
	}

	if q.len == 0 {
		q.lock = 0
		return nil
	}

	v := q.head.Value

	q.lock = 0
	return v
}

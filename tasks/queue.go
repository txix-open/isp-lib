package tasks

import (
	"runtime"
	"sync"
)

var (
	defaultConcurrentConsumers = runtime.NumCPU() + 1
	defaultBuffer              = 0
)

type Queue struct {
	c         chan interface{}
	f         func(value interface{})
	consumers int
	running   bool
	lock      sync.RWMutex
}

func (q *Queue) StartConsuming(f func(value interface{})) {
	q.f = f

	if !q.running {
		q.running = true
		q.runConsumers()
	}
}

func (q *Queue) Close() {
	close(q.c)
	q.running = false
}

func (q *Queue) Offer(value interface{}) {
	q.c <- value
}

func (q *Queue) Take() (interface{}, bool) {
	value, open := <-q.c
	return value, open
}

func (q *Queue) runConsumers() {
	for i := 0; i < q.consumers; i++ {
		go func() {
			for {
				value, open := q.Take()
				if !open {
					break
				}
				q.f(value)
			}
		}()
	}
}

func NewInMemoryQueue(concurrentConsumers, queueSize int) *Queue {
	if concurrentConsumers <= 0 {
		concurrentConsumers = defaultConcurrentConsumers
	}
	if queueSize <= 0 {
		queueSize = defaultBuffer
	}
	return &Queue{
		c:         make(chan interface{}, queueSize),
		consumers: concurrentConsumers,
	}
}

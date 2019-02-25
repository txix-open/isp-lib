package utils

import "sync/atomic"

type AtomicInt struct {
	v *int32
}

func (b *AtomicInt) Get() int {
	return int(atomic.LoadInt32(b.v))
}

func (b *AtomicInt) Set(value int) {
	atomic.StoreInt32(b.v, int32(value))
}

func (b *AtomicInt) IncAndGet() int {
	return int(atomic.AddInt32(b.v, 1))
}

func (b *AtomicInt) DecAndGet() int {
	return int(atomic.AddInt32(b.v, -1))
}

func NewAtomicInt(init int) *AtomicInt {
	v := int32(init)
	b := &AtomicInt{&v}
	return b
}

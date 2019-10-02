package atomic

import "sync/atomic"

type AtomicBool struct {
	v *int32
}

func (b *AtomicBool) Get() bool {
	if atomic.LoadInt32(b.v) == 1 {
		return true
	}
	return false
}

func (b *AtomicBool) Set(value bool) {
	if value {
		atomic.StoreInt32(b.v, 1)
	} else {
		atomic.StoreInt32(b.v, 0)
	}
}

func NewAtomicBool(init bool) *AtomicBool {
	b := &AtomicBool{}
	v := int32(0)
	if init {
		v = int32(1)
	}
	b.v = &v
	return b
}

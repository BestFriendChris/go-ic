package atomic_bool

import "sync/atomic"

type AtomicBool struct {
	a *atomic.Bool
}

func NewGlobal() *AtomicBool {
	return &AtomicBool{a: globalAtomicBool}
}

func NewNullable() (*AtomicBool, *atomic.Bool) {
	ab := &atomic.Bool{}
	return &AtomicBool{a: ab}, ab
}

func (ab AtomicBool) Val() bool {
	return ab.a.Load()
}

func (ab AtomicBool) Set() (old bool) {
	return ab.a.Swap(true)
}

var globalAtomicBool *atomic.Bool

func init() {
	globalAtomicBool = &atomic.Bool{}
}

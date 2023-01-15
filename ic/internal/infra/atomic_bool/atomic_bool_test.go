package atomic_bool

import "testing"

func Test_Normal(t *testing.T) {
	atomicBool := NewGlobal()
	t.Cleanup(func() {
		// Ensure we don't pollute other tests
		globalAtomicBool.Store(false)
	})

	if atomicBool.Val() {
		t.Error("expected to start as false")
	}

	if atomicBool.Set() {
		t.Error("expected to return previous value on Set()")
	}

	if !atomicBool.Set() {
		t.Error("now should be already set")
	}

	if !atomicBool.Val() {
		t.Error("expected to end as true")
	}
}

func Test_Nullable(t *testing.T) {
	atomicBool, underlyingState := NewNullable()

	if underlyingState.Load() == true {
		t.Error("expected to start as false")
	}

	if atomicBool.Val() == true {
		t.Error("expected to start as false")
	}
	if atomicBool.Set() == true {
		t.Error("expected to return previous value on Set()")
	}

	if atomicBool.Set() == false {
		t.Error("now should be already set")
	}

	if atomicBool.Val() == false {
		t.Error("expected to end as true")
	}

	if underlyingState.Load() == false {
		t.Error("expected to end as true")
	}
}

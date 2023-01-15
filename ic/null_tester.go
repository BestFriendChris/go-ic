package ic

import (
	"fmt"
	"sync/atomic"
)

func NewNullTester() *NullTester {
	return &NullTester{UpdateCalled: &atomic.Bool{}}
}

// NullTester is useful to fake out testing.T in order to verify we handle failures correctly
type NullTester struct {
	Failed bool
	Exited bool
	Output []string

	IsUpdateEnabled bool
	UpdateCalled    *atomic.Bool
}

func (nt *NullTester) Reset() {
	nt.Output = []string{}
	nt.Failed = false
	nt.Exited = false
}

// Implements ic.TestFileUpdater

func (nt *NullTester) UpdateEnabled() bool {
	return nt.IsUpdateEnabled
}

func (nt *NullTester) Update(ic *IC, _ string) {
	if nt.UpdateCalled.Swap(true) {
		ic.t.Log(`IC: already updated a test file. Skipping update. Rerun tests to try again`)
		return
	}

	ic.t.Log(`IC: Updating test file. Rerun tests to verify`)
}

// Implements ic.Tester

func (nt *NullTester) Helper() {
	// nothing to do
}

func (nt *NullTester) Log(args ...any) {
	nt.Output = append(nt.Output, fmt.Sprintln(args...))
}

func (nt *NullTester) Logf(format string, args ...any) {
	nt.Output = append(nt.Output, fmt.Sprintf(format, args...))
}

func (nt *NullTester) Fail() {
	nt.Failed = true
}

func (nt *NullTester) FailNow() {
	nt.Fail()
	nt.Exited = true
}

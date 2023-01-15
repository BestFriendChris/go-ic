package ic

import "fmt"

// NullTester is useful to fake out testing.T in order to verify we handle failures correctly
type NullTester struct {
	Failed bool
	Exited bool
	Output []string
}

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

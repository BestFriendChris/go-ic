package ic

// Tester is just the parts of testing.TB that we actually use.
// See NullTester for handy implementation for tests
type Tester interface {
	Fail()
	FailNow()
	Log(args ...any)
	Logf(format string, args ...any)
	Helper()
}

// DebugStringer allows for exactly defining the debug string
// used in tests.
type DebugStringer interface {
	DebugString() string
}

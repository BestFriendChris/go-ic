package ic_test

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BestFriendChris/go-ic/ic"
	"github.com/BestFriendChris/go-ic/ic/internal/infra/cmd"
)

func TestIC_Expect_simple(t *testing.T) {
	c := ic.New(t)
	c.Print("foo")
	c.Expect(`foo`)
}

func TestIC_Expect_trimExpectation(t *testing.T) {
	c := ic.New(t)
	c.Print("foo\nbar")
	c.Expect(`
		foo
		bar`)
}

func TestIC_Expect_trimInput(t *testing.T) {
	c := ic.New(t)
	c.Print("\tfoo\n\tbar")
	c.Expect(`
		foo
		bar`)
}

func TestIC_Expect_clearOutputAfterExpect(t *testing.T) {
	c := ic.New(t)
	c.Print("foo")
	c.Expect(`foo`)

	c.Print("bar")
	c.Expect(`bar`)
}

func TestIC_Expect_fail(t *testing.T) {
	c, nt, _, _ := newNullable()
	c.Print("this will succeed")
	c.Expect("this will fail")

	if !nt.Failed {
		t.Error("Expected this to fail")
	}

	if !nt.Exited {
		t.Error("Expected this to have exited the test")
	}

	want := `
 got: "this will succeed"
want: "this will fail"`
	if len(nt.Output) != 1 {
		t.Fatalf("got %d elements, want 1 element in:\n%#v", len(nt.Output), nt.Output)
	}
	got := nt.Output[0]
	if got != want {
		t.Errorf("\ngot %v\n\nwant %v", got, want)
	}
}

func TestIC_ExpectAndContinue_fail(t *testing.T) {
	c, nt, _, _ := newNullable()
	c.Print("this will succeed")
	c.ExpectAndContinue("this will fail")

	if !nt.Failed {
		t.Error("Expected this to fail")
	}

	if nt.Exited {
		t.Error("Expected this to NOT have exited the test")
	}

	want := `
 got: "this will succeed"
want: "this will fail"`

	if len(nt.Output) != 1 {
		t.Fatalf("got %d elements, want 1 element in:\n%#v", len(nt.Output), nt.Output)
	}
	got := nt.Output[0]
	if got != want {
		t.Errorf("\ngot %v\n\nwant %v", got, want)
	}
}

func TestIC_Expect_failWithMultipleLines(t *testing.T) {
	c, nt, _, _ := newNullable()
	c.Println("this will")
	c.Println("succeed")
	c.Expect(`
			this will
			fail
			`)

	if !nt.Failed {
		t.Error("Expected this to fail")
	}

	want := `
--- Got
+++ Want
@@ -1,3 +1,3 @@
 this will
-succeed
+fail
 
`
	if len(nt.Output) != 1 {
		t.Fatalf("got %d elements, want 1 element in:\n%#v", len(nt.Output), nt.Output)
	}
	got := nt.Output[0]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestIC_Expect_whenEmptyLines_updateEnabled(t *testing.T) {
	c, nt, _, ofc := newNullable()
	ofc.FlagEnabled = true

	c.Println("this will fail")
	c.Expect(``)

	if !nt.Failed {
		t.Error("Expected this to fail")
	}

	want := `IC: Updating test file. Rerun tests to verify
`
	if len(nt.Output) != 2 {
		t.Fatalf("got %d elements, want 2 elements in:\n%#v", len(nt.Output), nt.Output)
	}
	got := nt.Output[1]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

func TestIC_Expect_whenEmptyLines_updateDisabled(t *testing.T) {
	c, nt, _, ofc := newNullable()
	ofc.FlagEnabled = false
	ofc.EnvEnabled = false

	c.Println("this will")
	c.Println("fail")
	c.Expect(``)

	if !nt.Failed {
		t.Error("Expected this to fail")
	}

	want := `IC: update is disabled. enable with "-test.icupdate" flag or set the IC_UPDATE env var to anything
`
	if len(nt.Output) != 2 {
		t.Fatalf("got %d elements, want 2 elements in:\n%#v", len(nt.Output), nt.Output)
	}
	got := nt.Output[1]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

func TestIC_Expect_whenEmptyLines_updatingTwice(t *testing.T) {
	c, nt, alreadySeen, ofc := newNullable()
	ofc.FlagEnabled = true

	if alreadySeen.Load() != false {
		t.Error("should have started not already seen")
	}

	c.Println("this will fail and update")
	c.Expect(``)

	if alreadySeen.Load() != true {
		t.Error("now should have been already seen")
	}

	want := `IC: Updating test file. Rerun tests to verify
`
	if len(nt.Output) != 2 {
		t.Fatalf("got %d elements, want 2 elements in:\n%#v", len(nt.Output), nt.Output)
	}
	got := nt.Output[1]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}

	nt.Reset()
	c.Println("this will fail as well but not update")
	c.Expect(``)

	if !nt.Failed {
		t.Error("Expected this to fail")
	}

	want = `IC: already updated a test file. Skipping update. Rerun tests to try again
`
	if len(nt.Output) != 2 {
		t.Fatalf("got %d elements, want 2 elements in:\n%#v", len(nt.Output), nt.Output)
	}
	got = nt.Output[1]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

func TestIC_PrintVals(t *testing.T) {
	c := ic.New(t)

	foo := 1
	c.PrintValWithName("foo", foo)

	bar := "hi\nthere"
	// Aliased for PrintValWithName
	c.PVWN("bar", bar)

	baz := struct {
		A float32
		b bool
	}{2.1, false}
	c.PVWN("baz", baz)

	// Anonymous struct
	c.PrintVals(struct{ A, ignored, B int }{1, 2, 999})

	// Named struct
	type testStruct struct {
		D, ignored, E string
	}
	// Aliased for PrintVals
	c.PV(testStruct{
		D:       "foo",
		ignored: "bar",
		E:       "baz",
	})

	c.Expect(`
			foo: 1
			bar: "hi\nthere"
			baz: struct { A float32; b bool }{A:2.1, b:false}
			A: 1
			B: 999
			testStruct.D: "foo"
			testStruct.E: "baz"
			`)
}

func TestIC_Replace(t *testing.T) {
	c := ic.New(t)

	c.Replace(`\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d-\d\d:\d\d`, "1970-01-01T00:00:00-00:00")

	c.PVWN("now", time.Now().Format(time.RFC3339))
	c.Expect(`
			now: "1970-01-01T00:00:00-00:00"
			`)

	c.PVWN("later", time.Now().Format(time.RFC3339))
	c.Expect(`
			later: "1970-01-01T00:00:00-00:00"
			`)
}

func TestIC_ClearReplace(t *testing.T) {
	c := ic.New(t)

	c.Replace(`foo`, "bar")

	c.PVWN("first", "foo-bar")
	c.Expect(`
			first: "bar-bar"
			`)

	c.ClearReplace()
	c.Replace(`bar`, "baz")

	c.PVWN("second", "foo-bar")
	c.Expect(`
			second: "foo-baz"
			`)

}

func TestIC_PrintSep(t *testing.T) {
	c := ic.New(t)

	tests := []struct {
		Name       string
		Have, Want int
	}{
		{`Simple add`, 1 + 2, 3},
		{`Simple subtract`, 10 - 3, 7},
	}
	c.PrintSep()
	for _, test := range tests {
		c.PV(test)
		c.PS() // Alias for PrintSep
	}
	c.Expect(`
		--------------------------------------------------------------------------------
		Name: "Simple add"
		Have: 3
		Want: 3
		--------------------------------------------------------------------------------
		Name: "Simple subtract"
		Have: 7
		Want: 7
		--------------------------------------------------------------------------------
		`)
}

func TestIC_replaceOnEmpty_1(t *testing.T) {
	t.Skip("example of updating the test file")
	c := ic.New(t)

	c.Replace(`\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d-\d\d:\d\d`, "1970-01-01T00:00:00-00:00")

	c.PVWN("foo", 1)
	c.PVWN("bar", time.Now().Format(time.RFC3339))

	c.Expect(``)
}

func TestIC_replaceOnEmpty_2(t *testing.T) {
	t.Skip("example of updating the test file")
	t.Run("First", func(t *testing.T) {
		c := ic.New(t)
		c.Print("foo", 1)
		c.Expect(``)
	})
	t.Run("Second", func(t *testing.T) {
		c := ic.New(t)
		c.Print("bar", 2)
		c.ExpectAndContinue(``)

		c.Print("baz", 2)
		c.ExpectAndContinue(``)
	})
}

// Example from README.md
func TestComplex(t *testing.T) {
	c := ic.New(t)

	_, _ = fmt.Fprintln(&c.Writer, "You can write to the Writer directly")

	c.PrintValWithName("PrintValWithName", "Simplifies outputing values")
	c.PVWN("PVWN", "is an alias for PrintValWithName")

	c.PrintVals(struct{ A, B, c string }{
		"anonymous structs",
		"call PrintValWithName for each key",
		"but only the exported ones",
	})

	type TestingStruct struct {
		D, E string
	}
	c.PV(TestingStruct{
		D: "Named structs work as well",
		E: "and PV is an alias for PrintVals",
	})

	c.PrintSep()
	c.Println("You can use PrintSep to visually distinguish sections.")
	c.Println("PS is an alias for PrintSep")
	c.PS()

	tests := []struct {
		Name       string
		Have, Want int
	}{
		{"Adding 1 + 2", 1 + 2, 3},
		{"Subtracting 10 - 3", 10 - 3, 7},
	}
	for _, test := range tests {
		c.PV(test)
		c.PS()
	}

	c.Println("You can also use Replace to run regexp.ReplaceAll on the input before comparison")
	c.Println("For example, this will normalize the current time to something predictable")
	c.Replace(`\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d-\d\d:\d\d`, "1970-01-01T00:00:00-00:00")
	c.PVWN("Time", time.Now().Format(time.RFC3339))

	c.PS()
	c.Println("You can also indent the expectation string.")
	c.Println("The shortest line (after removing the leading newline) is used to trim spaces")
	c.PS()

	c.Println("Whenever you want to update your expectation,")
	c.Println("simply remove all content in the string and run the tests again")
	c.Println("Only one test will be replaced at a time, so multiple runs may be required")
	c.PS()

	c.Println("Running ExpectAndContinue will call t.Fail and allow a failed test to continue")
	c.ExpectAndContinue(`
        You can write to the Writer directly
        PrintValWithName: "Simplifies outputing values"
        PVWN: "is an alias for PrintValWithName"
        A: "anonymous structs"
        B: "call PrintValWithName for each key"
        TestingStruct.D: "Named structs work as well"
        TestingStruct.E: "and PV is an alias for PrintVals"
        --------------------------------------------------------------------------------
        You can use PrintSep to visually distinguish sections.
        PS is an alias for PrintSep
        --------------------------------------------------------------------------------
        Name: "Adding 1 + 2"
        Have: 3
        Want: 3
        --------------------------------------------------------------------------------
        Name: "Subtracting 10 - 3"
        Have: 7
        Want: 7
        --------------------------------------------------------------------------------
        You can also use Replace to run regexp.ReplaceAll on the input before comparison
        For example, this will normalize the current time to something predictable
        Time: "1970-01-01T00:00:00-00:00"
        --------------------------------------------------------------------------------
        You can also indent the expectation string.
        The shortest line (after removing the leading newline) is used to trim spaces
        --------------------------------------------------------------------------------
        Whenever you want to update your expectation,
        simply remove all content in the string and run the tests again
        Only one test will be replaced at a time, so multiple runs may be required
        --------------------------------------------------------------------------------
        Running ExpectAndContinue will call t.Fail and allow a failed test to continue
        `)

	c.Println("Every time you run Expect or ExpectAndContinue, the Output is reset for more testing")
	c.Println("Replacements are not reset by default. In order to remove all replacements, call ClearReplace")
	c.ClearReplace()
	c.Println("Running Expect will call t.FailNow")

	c.Expect(`
        Every time you run Expect or ExpectAndContinue, the Output is reset for more testing
        Replacements are not reset by default. In order to remove all replacements, call ClearReplace
        Running Expect will call t.FailNow
        `)
}

func newNullable() (ic.IC, *ic.NullTester, *atomic.Bool, *cmd.OverridableFlagChecker) {
	fakeFs := makeFakeFs()
	return ic.NewNullable(&fakeFs)
}

func makeFakeFs() (fakeFs map[string]string) {
	_, fName, lineNo, _ := runtime.Caller(0)

	var sb strings.Builder
	for i := 0; i < lineNo+10; i++ {
		_, _ = fmt.Fprintf(&sb, "line %d: Expect(``)\n", i+1)
	}

	//fmt.Printf("the file \"%s:%d\":\n%s", fName, lineNo, sb.String())

	fakeFs = map[string]string{
		fName: sb.String(),
	}
	return
}

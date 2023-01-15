package ic_test

import (
	"github.com/BestFriendChris/go-ic/ic"
	"reflect"
	"testing"
	"time"
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
	c, nt := ic.NewNullable()
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
	c, nt := ic.NewNullable()
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
	c, nt := ic.NewNullable()
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
	c, nt := ic.NewNullable()
	nt.IsUpdateEnabled = true

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
	c, nt := ic.NewNullable()
	nt.IsUpdateEnabled = false

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
	c, nt := ic.NewNullable()
	nt.IsUpdateEnabled = true

	c.Println("this will fail and update")
	c.Expect(``)

	nt.Reset()
	c.Println("this will fail as well but not update")
	c.Expect(``)

	if !nt.Failed {
		t.Error("Expected this to fail")
	}

	want := `IC: already updated a test file. Skipping update. Rerun tests to try again
`
	if len(nt.Output) != 2 {
		t.Fatalf("got %d elements, want 2 elements in:\n%#v", len(nt.Output), nt.Output)
	}
	got := nt.Output[1]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

func TestIC_PrintVals(t *testing.T) {
	c := ic.New(t)

	foo := 1
	c.PrintValWithName("foo", foo)

	bar := "hi\nthere"
	// Aliased as for PrintValWithName
	c.PVwN("bar", bar)

	baz := struct {
		A float32
		b bool
	}{2.1, false}
	c.PVwN("baz", baz)

	// Anonymous struct
	c.PrintVals(struct{ A, ignored, B int }{1, 2, 999})

	// Named struct
	type testStruct struct {
		D, ignored, E string
	}
	// Aliased as for PrintVals
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

	c.PVwN("now", time.Now().Format(time.RFC3339))
	c.Expect(`
			now: "1970-01-01T00:00:00-00:00"
			`)

	c.PVwN("later", time.Now().Format(time.RFC3339))
	c.Expect(`
			later: "1970-01-01T00:00:00-00:00"
			`)
}

func TestIC_ClearReplace(t *testing.T) {
	c := ic.New(t)

	c.Replace(`foo`, "bar")

	c.PVwN("first", "foo-bar")
	c.Expect(`
			first: "bar-bar"
			`)

	c.ClearReplace()
	c.Replace(`bar`, "baz")

	c.PVwN("second", "foo-bar")
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

	c.PVwN("foo", 1)
	c.PVwN("bar", time.Now().Format(time.RFC3339))

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

package ic_test

import (
	"go-ic/ic"
	"reflect"
	"testing"
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

func TestIC_Expect_multiple(t *testing.T) {
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

	want := []string{`
got  "this will succeed"
want "this will fail"`}
	if !reflect.DeepEqual(nt.Output, want) {
		t.Errorf("got %v want %v", nt.Output, want)
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

	want := []string{`
got  "this will succeed"
want "this will fail"`}
	if !reflect.DeepEqual(nt.Output, want) {
		t.Errorf("got %v want %v", nt.Output, want)
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
--- Want
+++ Got
@@ -1,3 +1,3 @@
 this will
-fail
+succeed
 
`
	if len(nt.Output) != 1 {
		t.Fatalf("got %d elements, want 1 element in:\n%#v", len(nt.Output), nt.Output)
	}
	got := nt.Output[0]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestIC_PrintVals(t *testing.T) {
	c := ic.New(t)

	foo := 1
	c.PrintValWithName("foo", foo)

	bar := "hi\nthere"
	c.PrintValWithName("bar", bar)

	baz := struct {
		A float32
		b bool
	}{2.1, false}
	c.PrintValWithName("baz", baz)

	// Anonymous struct
	c.PrintVals(struct{ A, ignored, B int }{1, 2, 999})

	// Named struct
	type testStruct struct {
		D, ignored, E string
	}
	c.PrintVals(testStruct{
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
			D: "foo"
			E: "baz"
			`)
}

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

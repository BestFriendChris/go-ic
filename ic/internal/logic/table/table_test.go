package table

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

type TestTable[T any] struct {
	Name       string
	Have, Want T
}

func TestPrintTable(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		var sb strings.Builder

		tests := []TestTable[int]{
			{"1 + 2", 1 + 2, 3},
			{"10 - 3", 10 - 3, 7},
			{"◊ foo", 1, 2},
		}
		_ = PrintTable(&sb, tests)

		got := sb.String()
		want := `
   | Name     | Have | Want |
---+----------+------+------+
 1 | "1 + 2"  | 3    | 3    |
---+----------+------+------+
 2 | "10 - 3" | 7    | 7    |
---+----------+------+------+
 3 | "◊ foo"  | 1    | 2    |
---+----------+------+------+
`[1:]
		if got != want {
			t.Errorf("\ngot:\n%swant:\n%s", got, want)
		}
	})
	t.Run("No Rows", func(t *testing.T) {
		var sb strings.Builder

		tests := make([]TestTable[int], 0)
		_ = PrintTable(&sb, tests)

		got := sb.String()
		want := `
   | Name | Have | Want |
---+------+------+------+
`[1:]
		if got != want {
			t.Errorf("\ngot:\n%swant:\n%s", got, want)
		}
	})
	t.Run("Slice of pointers to structs", func(t *testing.T) {
		var sb strings.Builder

		tests := []*TestTable[int]{
			{"1 + 2", 1 + 2, 3},
			{"10 - 3", 10 - 3, 7},
		}
		err := PrintTable(&sb, tests)
		if err != nil {
			t.Fatal(err)
		}

		got := sb.String()
		want := `
   | Name     | Have | Want |
---+----------+------+------+
 1 | "1 + 2"  | 3    | 3    |
---+----------+------+------+
 2 | "10 - 3" | 7    | 7    |
---+----------+------+------+
`[1:]
		if got != want {
			t.Errorf("\ngot:\n%swant:\n%s", got, want)
		}
	})
}

func TestPrintTable_errorCases(t *testing.T) {
	verifyError := func(w io.Writer, val any, want string) {
		t.Helper()
		err := PrintTable(w, val)
		if err == nil {
			t.Fatal("Expected error")
		}
		got := err.Error()
		if got != want {
			t.Errorf("got error %q\nwant error %q", got, want)
		}
	}
	t.Run("Not a slice", func(t *testing.T) {
		val := 1
		verifyError(&strings.Builder{}, val, "must be a slice: got int")
	})
	t.Run("Not a slice of structs", func(t *testing.T) {
		val := []string{"oops"}
		verifyError(&strings.Builder{}, val, "must be a slice of structs: got slice of string")
	})
	t.Run("Writer closed", func(t *testing.T) {
		w := &failureWriter{"expected failure"}
		verifyError(w, []TestTable[int]{}, "expected failure")
	})
}

func Test_colWidths(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		data := [][]string{
			{"Name", "Have", "Want", "◊"},
			{"s", "this is super long", "1", ""},
			{"really long", "shorter", "2", ""},
		}
		got := colWidths(data)
		want := []int{
			len("really long"),
			len("this is super long"),
			len("Want"),
			1,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("\nhave: %v\nwant: %v", got, want)
		}
	})
	t.Run("empty data", func(t *testing.T) {
		data := make([][]string, 0)
		got := colWidths(data)
		if len(got) != 0 {
			t.Errorf("expected length to be 0: %#v", got)
		}
	})
}

func Test_colWidths_errorCases(t *testing.T) {
	t.Run("panic when some rows shorter than header", func(t *testing.T) {
		badData := [][]string{
			{"Name", "Have", "Want"},
			{"this", "is", "good"},
			{"too short"},
		}
		assertPanicsWithMessage(t, "bad row length; expect len(3) got len(1)", func() {
			_ = colWidths(badData)
		})
	})
	t.Run("panic when some rows longer than header", func(t *testing.T) {
		badData := [][]string{
			{"Name", "Have", "Want"},
			{"this", "is", "good"},
			{"this", "is", "too", "long"},
		}
		assertPanicsWithMessage(t, "bad row length; expect len(3) got len(4)", func() {
			_ = colWidths(badData)
		})
	})
}

func Test_stringifyTableValues(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		data := []TestTable[float32]{
			{"One", 1, 2},
			{"Two", 300000, 4.123},
		}
		got := stringifyTableValues(reflect.ValueOf(data))
		want := [][]string{
			{"Name", "Have", "Want"},
			{`"One"`, "1", "2"},
			{`"Two"`, "300000", "4.123"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("\nhave: %#v\nwant: %#v", got, want)
		}
	})
	t.Run("ignore unexported fields", func(t *testing.T) {
		data := []struct {
			Exported, notExported string
		}{
			{"include", "hide"},
		}
		got := stringifyTableValues(reflect.ValueOf(data))
		want := [][]string{
			{"Exported"},
			{`"include"`},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("\nhave: %#v\nwant: %#v", got, want)
		}
	})
	t.Run("enum values", func(t *testing.T) {
		data := []struct {
			Name string
			TEV  testEnum
		}{
			{"One", testEnumVal1},
			{"Two", testEnumVal2},
		}
		got := stringifyTableValues(reflect.ValueOf(data))
		want := [][]string{
			{"Name", "TEV"},
			{`"One"`, "testEnum.testEnumVal1"},
			{`"Two"`, "testEnum.testEnumVal2"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("\nhave: %#v\nwant: %#v", got, want)
		}
	})
	t.Run("Slice of pointers to structs", func(t *testing.T) {
		data := []*TestTable[float32]{
			{"One", 1, 2},
			{"Two", 300000, 4.123},
		}
		got := stringifyTableValues(reflect.ValueOf(data))
		want := [][]string{
			{"Name", "Have", "Want"},
			{`"One"`, "1", "2"},
			{`"Two"`, "300000", "4.123"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("\nhave: %#v\nwant: %#v", got, want)
		}
	})
	t.Run("struct with pointers", func(t *testing.T) {
		s := "foo"
		data := []struct {
			Name string
			A    *string
		}{
			{"One", &s},
			{"Two", nil},
		}
		got := stringifyTableValues(reflect.ValueOf(data))
		want := [][]string{
			{"Name", "A"},
			{`"One"`, `"foo"`},
			{`"Two"`, ""},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("\nhave: %#v\nwant: %#v", got, want)
		}
	})
}

func Test_stringifyTableValues_errorCases(t *testing.T) {
	t.Run("not a slice", func(t *testing.T) {
		assertPanicsWithMessage(t, "must be slice: got int", func() {
			v := 1
			_ = stringifyTableValues(reflect.ValueOf(v))
		})
	})
	t.Run("not a slice of structs", func(t *testing.T) {
		assertPanicsWithMessage(t, "must be slice of structs: got slice of string", func() {
			v := []string{"oops"}
			_ = stringifyTableValues(reflect.ValueOf(v))
		})
	})
}

func Test_addHeader(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		headers := []string{"Super Long Header", "short", "Medium"}
		widths := []int{17, 20, 6}
		output := make([]string, 2)
		addHeader(&output, headers, widths)

		wants := []string{
			"   | Super Long Header | short                | Medium |",
			"---+-------------------+----------------------+--------+",
		}
		for i, got := range output {
			want := wants[i]
			if got != want {
				t.Errorf("line %d:\n got: %s\nwant: %s", i, got, want)
			}
		}
	})

}
func Test_addHeader_errorCases(t *testing.T) {
	t.Run("output too small", func(t *testing.T) {
		output := make([]string, 0)
		assertPanicsWithMessage(t, "output too small for input; must be at least len(2)", func() {
			headers := []string{"Super Long Header", "short", "Medium"}
			widths := []int{17, 20, 6}
			addHeader(&output, headers, widths)
		})
	})
	t.Run("headers and widths not same length", func(t *testing.T) {
		headers := []string{"Super Long Header", "short", "Medium"}
		widths := []int{17}
		assertPanicsWithMessage(t, "headers (len 3) and widths (len 1) not same length", func() {
			output := make([]string, 2)
			addHeader(&output, headers, widths)
		})
	})
}

func Test_addRows(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		rows := [][]string{
			{"row 1", "col 2", "really long cell"},
			{"row 2", "other long cell", "col 3"},
		}
		widths := []int{17, 20, 16}
		output := makeOutputFromRows(rows)
		addRows(&output, rows, widths)

		wants := []string{
			" 1 | row 1             | col 2                | really long cell |",
			"---+-------------------+----------------------+------------------+",
			" 2 | row 2             | other long cell      | col 3            |",
			"---+-------------------+----------------------+------------------+",
		}
		for i, got := range output[2:] {
			want := wants[i]
			if got != want {
				t.Errorf("line %d:\n got: %s\nwant: %s", i, got, want)
			}
		}
	})

	tests := []struct {
		Name string
		n    int
		want string
	}{
		{"less than 10 rows", 9, " 9 |"},
		{"less than 100 rows", 99, "99 |"},
		{"less than 1000 rows", 999, "999|"},
		{"more than 1000 rows", 1001, "001|"},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Helper()
			rows := makeRowsN(tt.n)
			widths := colWidths(rows)
			output := makeOutputFromRows(rows)
			addRows(&output, rows, widths)
			lastRow := len(output) - 2
			got := output[lastRow][:4]

			if got != tt.want {
				t.Errorf("\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func Test_addRows_errorCases(t *testing.T) {
	rows := [][]string{
		{"row 1", "col 2", "really long cell"},
		{"row 2", "other long cell", "col 3"},
	}
	widths := colWidths(rows)
	output := makeOutputFromRows(rows)
	t.Run("output too small", func(t *testing.T) {
		output := make([]string, 3)
		assertPanicsWithMessage(t, "output too small for input; must be at least len(6)", func() {
			addRows(&output, rows, widths)
		})
	})
	t.Run("rows and widths not same length", func(t *testing.T) {
		widths := []int{100}
		assertPanicsWithMessage(t, "row[0] (len 3) and widths (len 1) not same length", func() {
			addRows(&output, rows, widths)
		})
	})
	t.Run("rows must all be same length", func(t *testing.T) {
		rows := [][]string{
			{"row 1", "col 2", "really long cell"},
			{"row 2", "other long cell", "col 3", "OOPS"},
		}
		assertPanicsWithMessage(t, "row[1] (len 4) and widths (len 3) not same length", func() {
			addRows(&output, rows, widths)
		})
	})
}

/********************************************************************************
test helpers
********************************************************************************/

type failureWriter struct {
	msg string
}

func (f failureWriter) Write(_ []byte) (n int, err error) {
	return -1, fmt.Errorf(f.msg)
}

func assertPanicsWithMessage(t *testing.T, msg string, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		if r == nil {
			t.Fatalf("The code did not panic")
		}
		if r != msg {
			t.Fatalf("\n got panic %q\nwant panic %q", r, msg)
		}
	}()
	f()
}

func makeRowsN(n int) [][]string {
	e := make([][]string, n)
	for i := 0; i < n; i++ {
		e[i] = []string{"foo"}
	}
	return e
}

func makeOutputFromRows(rows [][]string) []string {
	return make([]string, (len(rows)*2)+2)
}

type testEnum int

const (
	testEnumVal1 testEnum = iota
	testEnumVal2
)

func (t testEnum) String() string {
	switch t {
	case testEnumVal1:
		return "testEnum.testEnumVal1"
	case testEnumVal2:
		return "testEnum.testEnumVal2"
	default:
		panic("unknown testEnum")
	}
}

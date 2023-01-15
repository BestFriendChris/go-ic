package ic

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

func New(t testing.TB) IC {
	return IC{t: t}
}

func NewNullable() (IC, *NullTester) {
	nt := NullTester{}
	return IC{t: &nt}, &nt
}

type IC struct {
	t            Tester
	Writer       bytes.Buffer
	replacements []replacement
}

func (ic *IC) Print(output ...any) {
	_, err := fmt.Fprint(&ic.Writer, output...)
	if err != nil {
		ic.t.Log(err)
		ic.t.FailNow()
	}
}

func (ic *IC) Println(output ...any) {
	_, err := fmt.Fprintln(&ic.Writer, output...)
	if err != nil {
		ic.t.Log(err)
		ic.t.FailNow()
	}
}

func (ic *IC) Printf(format string, a ...any) {
	_, err := fmt.Fprintf(&ic.Writer, format, a...)
	if err != nil {
		ic.t.Log(err)
		ic.t.FailNow()
	}
}

func (ic *IC) Expect(want string) {
	ic.t.Helper()
	if !ic.expectAndLog(want) {
		ic.t.FailNow()
	}
}

func (ic *IC) ExpectAndContinue(want string) {
	ic.t.Helper()
	if !ic.expectAndLog(want) {
		ic.t.Fail()
	}
}

func (ic *IC) expectAndLog(want string) (isSame bool) {
	got := trim(ic.Writer.String())
	for _, rp := range ic.replacements {
		got = rp.replace(got)
	}
	trimmedWant := trim(want)
	isSame = got == trimmedWant
	if !isSame {
		if isMultiline(want) {
			diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(trimmedWant),
				B:        difflib.SplitLines(got),
				FromFile: "Want",
				FromDate: "",
				ToFile:   "Got",
				ToDate:   "",
				Context:  1,
			})
			ic.t.Logf("\n%s", diff)
		} else {
			ic.t.Logf("\ngot  %q\nwant %q", got, trimmedWant)
		}
	}
	ic.Writer.Truncate(0)
	return
}

func (ic *IC) PrintVals(val any) {
	valType := reflect.TypeOf(val)
	if valType.Kind() != reflect.Struct {
		ic.t.Helper()
		ic.t.Logf("PrintVals must be called with a struct. Got %v", valType.Kind())
		ic.t.FailNow()
	}

	s := reflect.ValueOf(val)
	for i := 0; i < valType.NumField(); i++ {
		field := valType.Field(i)
		if field.IsExported() {
			name := field.Name
			value := s.Field(i).Interface()
			ic.PrintValWithName(name, value)
		}
	}
}

func (ic *IC) PrintValWithName(name string, val any) {
	ic.Printf("%s: %#v\n", name, val)
}

func (ic *IC) Replace(regex string, repl string) {
	re, err := regexp.Compile(regex)
	if err != nil {
		ic.t.Helper()
		ic.t.Log(err)
		ic.t.FailNow()
	}
	ic.replacements = append(ic.replacements, replacement{re, []byte(repl)})
}

func (ic *IC) ClearReplace() {
	ic.replacements = ic.replacements[:0]
}

type replacement struct {
	re   *regexp.Regexp
	repl []byte
}

func (r replacement) replace(s string) string {
	return string(r.re.ReplaceAll([]byte(s), r.repl))
}

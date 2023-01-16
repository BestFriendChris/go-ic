package ic

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/BestFriendChris/go-ic/ic/internal/infra/cmd"
	"github.com/BestFriendChris/go-ic/ic/internal/logic/table"
	"github.com/pmezard/go-difflib/difflib"
)

func New(t testing.TB) IC {
	return IC{t: t, testFileUpdater: NewTestFileUpdater()}
}

func NewNullable(testFiles *map[string]string) (IC, *NullTester, *atomic.Bool, *cmd.OverridableFlagChecker) {
	nt := NewNullTester()
	tfu, underlyingBool, ofc := NewNullableTestFileUpdater(testFiles)
	return IC{t: nt, testFileUpdater: tfu}, nt, underlyingBool, ofc
}

// IC is the test value runner. Create with New(*testing.TB)
type IC struct {
	t               Tester
	Writer          bytes.Buffer
	replacements    []replacement
	testFileUpdater TestFileUpdater
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

// Expect will compare the provided string to all the calls to ic.Print*
// combined. If "want" is an empty string, the library will automatically replace
// it with the provided value if either is set:
//   - "IC_UPDATE" environment variable
//   - "-test.icupdate" command line flag is set
//
// Expect will fail the test immediately on failure. ExpectAndContinue can be
// used to keep running the rest of the test
func (ic *IC) Expect(want string) {
	ic.t.Helper()
	if !ic.expectAndLog(want) {
		ic.t.FailNow()
	}
}

// ExpectAndContinue behaves exactly like Expect, with the exception that the
// test will continue to run on a failure
func (ic *IC) ExpectAndContinue(want string) {
	ic.t.Helper()
	if !ic.expectAndLog(want) {
		ic.t.Fail()
	}
}

func (ic *IC) expectAndLog(want string) (isSame bool) {
	ic.t.Helper()
	got := trim(ic.Writer.String())
	for _, rp := range ic.replacements {
		got = rp.replace(got)
	}
	isSame = ic.logDiffIfDifferent(want, got)
	ic.Writer.Truncate(0)
	if len(want) == 0 {
		if ic.testFileUpdater.UpdateEnabled() {
			ic.testFileUpdater.Update(ic, got)
			return false
		} else {
			ic.t.Log(`IC: update is disabled. enable with "-test.icupdate" flag or set the IC_UPDATE env var to anything`)
		}
	}
	return
}

func (ic *IC) logDiffIfDifferent(want string, got string) (isSame bool) {
	ic.t.Helper()
	trimmedWant := trim(want)
	isSame = got == trimmedWant
	if !isSame {
		if isMultiline(want) || isMultiline(got) {
			diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(got),
				B:        difflib.SplitLines(trimmedWant),
				FromFile: "Got",
				FromDate: "",
				ToFile:   "Want",
				ToDate:   "",
				Context:  3,
			})
			ic.t.Logf("\n%s", diff)
		} else {
			ic.t.Logf("\n got: %q\nwant: %q", got, trimmedWant)
		}
	}
	return
}

// TT is a test table struct for PrintTable or PrintVals
type TT[T any] struct {
	Name       string
	Have, Want T
}

// PT is an alias for PrintTable
func (ic *IC) PT(val any) {
	ic.t.Helper()
	ic.PrintTable(val)
}

// PrintTable will take an array of structs and print a table
func (ic *IC) PrintTable(val any) {
	ic.t.Helper()
	err := table.PrintTable(&ic.Writer, val)
	if err != nil {
		ic.t.Logf("PrintTable: %s", err)
		ic.t.FailNow()
	}
}

// PV is an alias for PrintVals
func (ic *IC) PV(val any) {
	ic.t.Helper()
	ic.PrintVals(val)
}

// PrintVals will take any struct and call PrintValWithName on each of the exported fields
func (ic *IC) PrintVals(val any) {
	ic.t.Helper()
	valType := reflect.TypeOf(val)
	if valType.Kind() != reflect.Struct {
		ic.t.Logf("PrintVals must be called with a struct. Got %v", valType.Kind())
		ic.t.FailNow()
	}

	s := reflect.ValueOf(val)
	for i := 0; i < valType.NumField(); i++ {
		field := valType.Field(i)
		if field.IsExported() {
			name := field.Name
			if valType.Name() != "" {
				name = fmt.Sprintf("%s.%s", valType.Name(), name)
			}
			value := s.Field(i).Interface()
			ic.PrintValWithName(name, value)
		}
	}
}

// PVWN is an alias for PrintValWithName
func (ic *IC) PVWN(name string, val any) {
	ic.PrintValWithName(name, val)
}

// PrintValWithName is a simple formatter for testing values
func (ic *IC) PrintValWithName(name string, val any) {
	ic.Printf("%s: %#v\n", name, val)
}

func (ic *IC) PS() {
	ic.PrintSep()
}

var separator = strings.Repeat("-", 80)

func (ic *IC) PrintSep() {
	ic.Println(separator)
}

// Replace can be used to run a regexp.ReplaceAll on the output before comparison
func (ic *IC) Replace(regex string, repl string) {
	ic.t.Helper()
	re, err := regexp.Compile(regex)
	if err != nil {
		ic.t.Log(err)
		ic.t.FailNow()
	}
	ic.replacements = append(ic.replacements, replacement{re, []byte(repl)})
}

// ClearReplace can be used to reset the active replacements
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

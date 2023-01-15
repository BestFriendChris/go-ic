# go-ic (go Instant Camera)

IC is a snapshot testing framework. It has a record 
and playback system with optional automatic updating. This was inspired
by the article: [Jane Street - What if writing tests was a joyful experience?](https://blog.janestreet.com/the-joy-of-expect-tests/)

## Simple Usage

Create a test with an empty `c.Expect("")`.

```go
package example_test

import (
    "github.com/BestFriendChris/go-ic/ic"
    "testing"
)

func TestExample(t *testing.T) {
    c := ic.New(t)
    c.Println("foo")
    c.Expect(``)
}
```

Run your tests with either `IC_UPDATE=1` or the param
`-test.icupdate`

```shell
$ IC_UPDATE=1 go test ./...
--- FAIL: TestExample (0.00s)
    example_test.go:8:
        --- Got
        +++ Want
        @@ -1,2 +1 @@
        -foo

    ic_test.go:331: IC: Updating test file. Rerun tests to verify
```

The test still fails after updating. Rerun the tests
to verify it worked

## Complex Example

```go
func TestComplex(t *testing.T) {
    c := ic.New(t)
    
    fmt.Fprintln(&c.Writer, "You can write to the Writer directly")
    
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

```
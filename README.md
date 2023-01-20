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

    c.PrintSep()
    c.Println("You can use PrintSep to visually distinguish sections.")
    c.PS()
    c.Println("PS is an alias for PrintSep")

    c.Println()
    c.PrintSection("c.PrintSection")
    c.Println("Create a section header with PrintSection")

    c.Println()
    c.PrintSection("c.Writer")
    _, _ = fmt.Fprintln(&c.Writer, "You can write to the Writer directly")

    c.Println()
    c.PrintSection("c.PrintValWithName (alias c.PVWN)")

    c.PrintValWithName("PrintValWithName", "Simplifies outputting values")
    c.PVWN("PVWN", "is an alias for PrintValWithName")

    c.Println()
    c.PrintSection("c.PrintVals (alias c.PV)")

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

    c.Println()
    c.PrintSection("c.PrintTable (alias c.PT)")

    c.Println("You can print an array of structs as a table as well with PrintTable (or PT)")
    c.PrintTable([]TestingStruct{
        {"r1c1", "r1c2"},
        {"r2c1", "r2c2"},
    })

    c.Println()
    c.PrintSection("ic.TT")

    c.Println("ic.TT is a pre-made struct for PrintTable and PrintVals")
    tt := []ic.TT[int]{
        {"Adding 1 + 2", 1 + 2, 3},
        {"Subtracting 10 - 3", 10 - 3, 7},
    }
    c.PT(tt)

    c.Println()
    c.PrintSection("c.PrintVals with a test table")

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

    c.Println()
    c.PrintSection("c.Replace")

    c.Println("You can also use Replace to run regexp.ReplaceAll on the input before comparison")
    c.Println("For example, this will normalize the current time to something predictable")
    c.Replace(`\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d-\d\d:\d\d`, "1970-01-01T00:00:00-00:00")
    c.PVWN("Time", time.Now().Format(time.RFC3339))

    c.Println()
    c.PrintSection("Indentation")

    c.Println("You can also indent the expectation string.")
    c.Println("The shortest line (after removing the leading newline) is used to trim spaces")

    c.Println()
    c.PrintSection("Updating")

    c.Println("Whenever you want to update your expectation,")
    c.Println("simply remove all content in the string and run the tests again")
    c.Println("Only one test will be replaced at a time, so multiple runs may be required")

    c.Println()
    c.PrintSection("ExpectAndContinue")

    c.Println("Running ExpectAndContinue will call t.Fail and allow a failed test to continue")
    c.ExpectAndContinue(`
        --------------------------------------------------------------------------------
        You can use PrintSep to visually distinguish sections.
        --------------------------------------------------------------------------------
        PS is an alias for PrintSep
        
        ################################################################################
        # c.PrintSection
        ################################################################################
        Create a section header with PrintSection
        
        ################################################################################
        # c.Writer
        ################################################################################
        You can write to the Writer directly
        
        ################################################################################
        # c.PrintValWithName (alias c.PVWN)
        ################################################################################
        PrintValWithName: "Simplifies outputting values"
        PVWN: "is an alias for PrintValWithName"
        
        ################################################################################
        # c.PrintVals (alias c.PV)
        ################################################################################
        A: "anonymous structs"
        B: "call PrintValWithName for each key"
        TestingStruct.D: "Named structs work as well"
        TestingStruct.E: "and PV is an alias for PrintVals"
        
        ################################################################################
        # c.PrintTable (alias c.PT)
        ################################################################################
        You can print an array of structs as a table as well with PrintTable (or PT)
           | D      | E      |
        ---+--------+--------+
         1 | "r1c1" | "r1c2" |
        ---+--------+--------+
         2 | "r2c1" | "r2c2" |
        ---+--------+--------+
        
        ################################################################################
        # ic.TT
        ################################################################################
        ic.TT is a pre-made struct for PrintTable and PrintVals
           | Name                 | Have | Want |
        ---+----------------------+------+------+
         1 | "Adding 1 + 2"       | 3    | 3    |
        ---+----------------------+------+------+
         2 | "Subtracting 10 - 3" | 7    | 7    |
        ---+----------------------+------+------+
        
        ################################################################################
        # c.PrintVals with a test table
        ################################################################################
        Name: "Adding 1 + 2"
        Have: 3
        Want: 3
        --------------------------------------------------------------------------------
        Name: "Subtracting 10 - 3"
        Have: 7
        Want: 7
        --------------------------------------------------------------------------------
        
        ################################################################################
        # c.Replace
        ################################################################################
        You can also use Replace to run regexp.ReplaceAll on the input before comparison
        For example, this will normalize the current time to something predictable
        Time: "1970-01-01T00:00:00-00:00"
        
        ################################################################################
        # Indentation
        ################################################################################
        You can also indent the expectation string.
        The shortest line (after removing the leading newline) is used to trim spaces
        
        ################################################################################
        # Updating
        ################################################################################
        Whenever you want to update your expectation,
        simply remove all content in the string and run the tests again
        Only one test will be replaced at a time, so multiple runs may be required
        
        ################################################################################
        # ExpectAndContinue
        ################################################################################
        Running ExpectAndContinue will call t.Fail and allow a failed test to continue
        `)

    c.Println("Every time you run Expect or ExpectAndContinue, the Output is reset for more testing")
    c.Println()
    c.PrintSection("c.ClearReplace")
    c.Println("Replacements are not reset by default. In order to remove all replacements, call ClearReplace")
    c.ClearReplace()

    c.Println()
    c.PrintSection("c.Expect")
    c.Println("Running Expect will call t.FailNow")

    c.Expect(`
        Every time you run Expect or ExpectAndContinue, the Output is reset for more testing
        
        ################################################################################
        # c.ClearReplace
        ################################################################################
        Replacements are not reset by default. In order to remove all replacements, call ClearReplace
        
        ################################################################################
        # c.Expect
        ################################################################################
        Running Expect will call t.FailNow
        `)
}
```
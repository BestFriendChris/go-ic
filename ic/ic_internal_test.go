package ic

import "testing"

func Test_trim(t *testing.T) {
	t.Run("no newlines", func(t *testing.T) {
		assertEqual(t, trim(`foo`), `foo`)
		assertEqual(t, trim(`  foo`), `  foo`)
		assertEqual(t, trim(`foo  `), `foo  `)
		assertEqual(t, trim("\tfoo"), "\tfoo")
		assertEqual(t, trim("foo\t"), "foo\t")
	})
	t.Run("newlines at beginning", func(t *testing.T) {
		input := `
foo`
		assertEqual(t, trim(input), "foo")
	})
	t.Run("leading spaces", func(t *testing.T) {
		tests := []struct {
			want, input string
		}{
			{"foo\nbar", `
	foo
	bar`},
			{"foo\nbar", `
  foo
  bar`},
			{"foo\n\tbar\nbaz", `
	foo
		bar
	baz`},
			{"\tfoo\n\tbar\n", `
	foo
	bar
`},
		}
		for _, tt := range tests {
			assertEqual(t, trim(tt.input), tt.want)
		}

	})
}

func assertEqual(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

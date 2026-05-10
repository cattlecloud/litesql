package litesql

import (
	"testing"

	"github.com/shoenig/test/must"
)

func Test_SanitizeFTS5(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		exp   string
	}{
		{"asterisk removed", "foo*bar", "foobar"},
		{"colon removed", "foo:bar", "foobar"},
		{"caret removed", "foo^bar", "foobar"},
		{"quote removed", `foo"bar`, "foobar"},
		{"acceptance", "hello world:foo*bar ^baz", "hello worldfoobar baz"},
		{"empty string", "", ""},
		{"no sanitization needed", "simple query", "simple query"},
		{"multiple spaces normalized", "foo  bar   baz", "foo bar baz"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeFTS5(tc.input)
			must.Eq(t, tc.exp, result)
		})
	}
}

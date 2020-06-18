package search

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testToken struct {
	token rune
	val   string
}

func TestLexer(t *testing.T) {
	testcases := []struct {
		input  string
		tokens []testToken
		err    string
	}{
		{
			input: "foo bar",
			tokens: []testToken{
				{token: tokTerm, val: "foo"},
				{token: tokTerm, val: "bar"},
			},
		}, {
			input: " foo ",
			tokens: []testToken{
				{token: tokTerm, val: "foo"},
			},
		}, {
			input: "field:value",
			tokens: []testToken{
				{token: tokTerm, val: "field"},
				{token: ':', val: ":"},
				{token: tokTerm, val: "value"},
			},
		}, {
			input: `\\ field\:value ryan\'s\ hat\`,
			tokens: []testToken{
				{token: tokTerm, val: `\`},
				{token: tokTerm, val: `field:value`},
				{token: tokTerm, val: `ryan's hat`},
			},
		}, {
			input: "\\\U0001f600",
			tokens: []testToken{
				{token: tokTerm, val: "\U0001f600"},
			},
		}, {
			input: "`foo`",
			tokens: []testToken{
				{token: '`', val: "`"},
				{token: tokTerm, val: "foo"},
				{token: '`', val: "`"},
			},
		},
	}
	for i, c := range testcases {
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			var val string
			result := make([]testToken, 0)
			l := newLexer(c.input)
			for {
				tok := l.Next(&val)
				if tok == tokEOF {
					break
				} else if tok == tokError {
					require.Equal(t, c.err, val, "input: %v", c.input)
					return
				}
				result = append(result, testToken{token: tok, val: val})
			}
			require.Equal(t, c.tokens, result, "input: %v", c.input)
		})
	}
}

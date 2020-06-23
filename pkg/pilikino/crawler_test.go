package pilikino

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveLink(t *testing.T) {
	index, err := NewMemOnlyIndex()
	require.NoError(t, err)
	absPath, err := filepath.Abs("crawler_test.go")
	require.NoError(t, err)

	cases := []struct {
		from, to string
		result   string
		err      string
	}{
		{
			from:   "crawler_test.go",
			to:     "parser_test",
			result: "parser_test.go",
		}, {
			from: "crawler_test.go",
			to:   "_test",
			err:  ErrAmbiguousLink.Error(),
		}, {
			from: "crawler_test.go",
			to:   "INVALID LINK",
			err:  ErrDeadLink.Error(),
		}, {
			from:   absPath,
			to:     "parser_test.go",
			result: "parser_test.go",
		}, {
			from:   "",
			to:     absPath,
			result: "crawler_test.go",
		}, {
			from: "hostile.md",
			to:   "foo/../../../../../../../etc/passwd",
			err:  "invalid relative link",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			result, err := index.ResolveLink(c.from, c.to)
			if c.err != "" {
				require.Equal(t, c.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.result, result)
			}
		})
	}
}

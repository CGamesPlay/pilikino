package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type equalityError struct {
	err  error
	path []string
}

func errorAtPath(root string, err error) *equalityError {
	if ee, ok := err.(*equalityError); ok {
		path := make([]string, 1, len(ee.path)+1)
		path[0] = root
		path = append(path, ee.path...)
		return &equalityError{
			err:  ee.err,
			path: path,
		}
	}
	return &equalityError{
		err:  err,
		path: []string{root},
	}
}

func (e *equalityError) Error() string {
	if len(e.path) > 0 {
		return fmt.Sprintf("%v at %v", e.err, strings.Join(e.path, "."))
	}
	return e.err.Error()
}

func matchPart(expected, actual interface{}) error {
	switch ex := expected.(type) {
	case map[string]interface{}:
		ac, ok := actual.(map[string]interface{})
		if !ok {
			return errors.New("not a map")
		}
		for k, exV := range ex {
			if acV, ok := ac[k]; !ok {
				return fmt.Errorf("missing key %#v", k)
			} else if err := matchPart(exV, acV); err != nil {
				return errorAtPath(k, err)
			}
		}
		return nil
	case []interface{}:
		ac, ok := actual.([]interface{})
		if !ok {
			return errors.New("not an array")
		}
		if len(ac) != len(ex) {
			return errors.New("different lengths")
		}
		for i := range ex {
			if err := matchPart(ex[i], ac[i]); err != nil {
				return errorAtPath(fmt.Sprintf("[%d]", i), err)
			}
		}
		return nil
	default:
		if !assert.ObjectsAreEqual(expected, actual) {
			return errors.New("not equal")
		}
		return nil
	}
}

func requireMatch(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	// If the pieces don't match, perform a normal equality comparison which
	// will fail and show the full diff. This is similar to the requireMatch
	// matcher in jest.
	if err := matchPart(expected, actual); err != nil {
		msg := errorAtPath("actual", err).Error()
		if len(msgAndArgs) > 0 {
			msg += "\n" + fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		require.Equal(t, expected, actual, msg)
		panic("matchPart returned false") // but require.Equal didn't fail!
	}
}

func TestParser(t *testing.T) {
	testcases := []struct {
		input  string
		result map[string]interface{}
		err    string
	}{
		{
			input: "foo",
			result: map[string]interface{}{
				"should": map[string]interface{}{"disjuncts": []interface{}{
					map[string]interface{}{"match": "foo"},
				}},
			},
		}, {
			input: "field:foo",
			result: map[string]interface{}{
				"should": map[string]interface{}{"disjuncts": []interface{}{
					map[string]interface{}{"field": "field", "match": "foo"},
				}},
			},
		}, {
			input: "foo bar",
			result: map[string]interface{}{
				"should": map[string]interface{}{"disjuncts": []interface{}{
					map[string]interface{}{"match": "foo"},
					map[string]interface{}{"match": "bar"},
				}},
			},
		}, {
			input: `"foo bar"`,
			result: map[string]interface{}{
				"should": map[string]interface{}{"disjuncts": []interface{}{
					map[string]interface{}{"match_phrase": "foo bar"},
				}},
			},
		}, {
			input: "`literal text`",
			result: map[string]interface{}{
				"should": map[string]interface{}{"disjuncts": []interface{}{
					map[string]interface{}{
						"terms": []interface{}{"literal", "text"},
					},
				}},
			},
		}, {
			input: "#tag",
			result: map[string]interface{}{
				"should": map[string]interface{}{"disjuncts": []interface{}{
					map[string]interface{}{"field": "tags", "match": "tag"},
				}},
			},
		}, {
			input: "links:20200521-1059-git_blame.md",
			result: map[string]interface{}{
				"should": map[string]interface{}{"disjuncts": []interface{}{
					map[string]interface{}{"field": "links", "match": "20200521-1059-git_blame.md"},
				}},
			},
		},
	}
	for i, c := range testcases {
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			l := newLexer(c.input)
			yyParse(l)
			if c.err != "" {
				require.Equal(t, c.err, l.err, "input: %v", c.input)
			} else {
				require.NoError(t, l.err, "input: %v\npos: %d", c.input, l.start)
				queryJSON, err := json.Marshal(l.result)
				require.NoError(t, err, "input: %v", c.input)
				var restoredQuery map[string]interface{}
				err = json.Unmarshal(queryJSON, &restoredQuery)
				require.NoError(t, err, "input: %v", c.input)
				requireMatch(t, c.result, restoredQuery, "input: %v", c.input)
			}
		})
	}
}

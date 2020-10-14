package search

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/blevesearch/bleve/search/query"
)

//go:generate go run golang.org/x/tools/cmd/goyacc -l -o query_parser.go query_parser.y

// specialChars are never part of a term, unless prefixed by a backslash
const specialChars = "`:\"*"

// isTermRune checks if a character can be part of a term at the given index
func isTermRune(r rune, pos int) bool {
	if pos == 0 {
		return unicode.IsGraphic(r) && !unicode.IsPunct(r) && !unicode.IsSpace(r) && strings.IndexRune(specialChars, r) == -1
	}
	return unicode.IsGraphic(r) && !unicode.IsSpace(r) && strings.IndexRune(specialChars, r) == -1
}

func init() {
	yyErrorVerbose = true
}

type lexer struct {
	input  string
	start  int
	pos    int
	width  int
	err    error
	result query.Query
}

func newLexer(input string) *lexer {
	return &lexer{input: input}
}

func (l *lexer) Next(val *string) rune {
ignore:
	r := l.peek()
	switch {
	case unicode.IsSpace(r):
		l.nextRune()
		l.ignore()
		goto ignore
	case isTermRune(r, 0) || r == '\\':
		return l.nextTerm(val)
	case unicode.IsGraphic(r):
		l.nextRune()
		*val = l.emit()
		return r
	case r == tokEOF:
		*val = ""
		return tokEOF
	default:
		*val = fmt.Sprintf("unexpected U+%04x at %d", r, l.start)
		return tokError
	}
}

func (l *lexer) Lex(tok *yySymType) int {
	r := int(l.Next(&tok.term))
	if r == tokEOF {
		return 0
	}
	return r
}

func (l *lexer) Error(err string) {
	l.err = errors.New(err)
}

func (l *lexer) nextTerm(val *string) rune {
	hasSlashes := false
	for {
		r := l.nextRune()
		if r == '\\' {
			// swallow and continue
			hasSlashes = true
			l.nextRune()
			continue
		} else if !isTermRune(r, l.pos-l.start-1) {
			l.backup()
			break
		}
	}
	if !hasSlashes {
		// Most terms don't have slashes, so skip a needless scan+copy
		*val = l.emit()
		return tokTerm
	}
	escaped := l.emit()
	result := strings.Builder{}
	// unescape backslashes
	for len(escaped) > 0 {
		slash := strings.IndexRune(escaped, '\\')
		if slash == -1 {
			// No slashes, write rest of string
			result.WriteString(escaped)
			break
		} else if slash == len(escaped)-1 {
			// Write rest of string ignoring trailing backslash
			result.WriteString(escaped[0:slash])
			break
		}

		r, w := utf8.DecodeRuneInString(escaped[slash+1:])
		result.WriteString(escaped[0:slash])
		result.WriteRune(r)
		escaped = escaped[slash+1+w:]
	}
	*val = result.String()
	return tokTerm
}

// next returns the next rune in the input.
func (l *lexer) nextRune() (r rune) {
	if l.pos >= len(l.input) {
		l.width = ^0
		return tokEOF
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	if l.width == ^0 {
		// "put back" the eof
		return
	} else if l.width == 0 {
		panic("invalid backup")
	}
	l.pos -= l.width
	l.width = 0
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
	r := l.nextRune()
	l.backup()
	return r
}

// emit stores the value of the token and advances
func (l *lexer) emit() string {
	start := l.start
	l.start = l.pos
	return l.input[start:l.start]
}

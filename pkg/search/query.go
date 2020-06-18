package search

import (
	"fmt"

	"github.com/blevesearch/bleve/search/query"
)

// ParseQuery parses a query string and returns a bleve.Query. The parsing is
// designed to be permissive, but some combinations are still undecipherable,
// and so an error will be returned in these cases.
func ParseQuery(query string) (query.Query, error) {
	l := newLexer(query)
	yyParse(l)
	if l.err != nil {
		return nil, l.err
	}
	return l.result, nil
}

func setResult(l yyLexer, v query.Query) {
	l.(*lexer).result = v
}

func getMulti(q query.Query) []query.Query {
	switch j := q.(type) {
	case *query.ConjunctionQuery:
		return j.Conjuncts
	case *query.DisjunctionQuery:
		return j.Disjuncts
	default:
		return []query.Query{q}
	}
}

func mergeMulti(a, b query.Query) []query.Query {
	if a == nil && b == nil {
		return nil
	}
	var aQueries, bQueries []query.Query
	length := 0
	if a != nil {
		aQueries = getMulti(a)
		length += len(aQueries)
	}
	if b != nil {
		bQueries = getMulti(b)
		length += len(bQueries)
	}
	merged := make([]query.Query, 0, length)
	if aQueries != nil {
		merged = append(merged, aQueries...)
	}
	if bQueries != nil {
		merged = append(merged, bQueries...)
	}
	return merged
}

func mergeBoolean(a, b *query.BooleanQuery) *query.BooleanQuery {
	must := mergeMulti(a.Must, b.Must)
	should := mergeMulti(a.Should, b.Should)
	mustNot := mergeMulti(a.MustNot, b.MustNot)
	return query.NewBooleanQueryForQueryString(must, should, mustNot)
}

// It turns out a bunch of stuff in bleve doesn't actually implement
// FieldableQuery
func setField(q query.Query, field string) query.Query {
	switch fq := q.(type) {
	case query.FieldableQuery:
		fq.SetField(field)
	case *query.PhraseQuery:
		fq.Field = field
	default:
		panic(fmt.Sprintf("cannot re-field %#v", fq))
	}
	return q
}

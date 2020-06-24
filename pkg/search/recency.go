package search

import (
	"math"
	"time"

	"github.com/blevesearch/bleve/index"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/numeric"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/blevesearch/bleve/search/searcher"
)

// RecencyQuery wraps another query and rescores it based on the recency of a
// specific field.
type RecencyQuery struct {
	// BoostVal represents the ratio of recency to preexisitng score. The
	// default, 1.0, assigns equal importance to recency score and match score.
	// A value of 2 would relatively rank recency twice as important as match
	// score.
	BoostVal *query.Boost
	// Field is the ranked field
	Field string
	// RecentAge is the age at which the score reaches BoostVal / 2. The
	// default value is 7 days.
	RecentAge time.Duration
	base      query.Query
}

// NewRecencyQuery takes a base query, but mixes in a score based on how recent
// the passed field is.
func NewRecencyQuery(field string, base query.Query) *RecencyQuery {
	return &RecencyQuery{Field: field, RecentAge: time.Hour * 24 * 7, base: base}
}

// SetBoost satisfies query.BoostableQuery
func (q *RecencyQuery) SetBoost(b float64) {
	boost := query.Boost(b)
	q.BoostVal = &boost
}

// Boost satisfies query.BoostableQuery
func (q *RecencyQuery) Boost() float64 {
	return q.BoostVal.Value()
}

// Searcher satisfied query.Query.
func (q *RecencyQuery) Searcher(i index.IndexReader, m mapping.IndexMapping, options search.SearcherOptions) (search.Searcher, error) {
	bs, err := q.base.Searcher(i, m, options)
	if err != nil {
		return nil, err
	}

	dvReader, err := i.DocValueReader([]string{q.Field})
	if err != nil {
		return nil, err
	}

	return searcher.NewFilteringSearcher(bs, buildRecencyFilter(dvReader, q.Field, q.RecentAge, q.BoostVal.Value())), nil
}

func buildRecencyFilter(dvReader index.DocValueReader, field string, recentAge time.Duration, boost float64) searcher.FilterFunc {
	return func(d *search.DocumentMatch) bool {
		// score against the newest date
		var selected int64

		err := dvReader.VisitDocValues(d.IndexInternalID, func(field string, term []byte) {
			// only consider the values which are shifted 0
			prefixCoded := numeric.PrefixCoded(term)
			shift, err := prefixCoded.Shift()
			if err == nil && shift == 0 {
				i64, err := prefixCoded.Int64()
				if err == nil {
					if i64 > selected {
						selected = i64
					}
				}
			}
		})
		if err != nil {
			return true
		}
		// Rescore based on date
		date := time.Unix(0, selected)
		age := time.Now().Sub(date)
		score := math.Min(1, math.Pow(2, -float64(age)/float64(recentAge)))
		d.Score = (d.Score + score*boost) / (boost + 1)
		return true
	}
}

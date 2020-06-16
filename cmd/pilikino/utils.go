package main

import (
	"os"
	"unicode"

	"github.com/CGamesPlay/pilikino/pkg/pilikino"
	"github.com/CGamesPlay/pilikino/pkg/search"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

func getIndex() (*pilikino.Index, error) {
	if directory != "" {
		if err := os.Chdir(directory); err != nil {
			return nil, err
		}
	}
	index, err := pilikino.NewMemOnlyIndex()
	if err != nil {
		return nil, err
	}
	return index, nil
}

func parseQuery(queryString string, includeAll bool) (query.Query, error) {
	var defaultMatch query.Query
	if includeAll {
		matchAll := query.NewMatchAllQuery()
		matchAll.SetBoost(0.1)
		defaultMatch = matchAll
	} else {
		defaultMatch = query.NewMatchNoneQuery()
	}
	var baseQuery *query.BooleanQuery
	if len(queryString) == 0 {
		baseQuery = query.NewBooleanQuery([]query.Query{defaultMatch}, nil, nil)
	} else {
		if runes := []rune(queryString); unicode.IsLetter(runes[len(runes)-1]) {
			queryString += "*"
		}
		parsed, err := bleve.NewQueryStringQuery(queryString).Parse()
		if err != nil {
			return nil, err
		}
		baseQuery = query.NewBooleanQuery([]query.Query{parsed}, nil, nil)
	}
	recency := search.NewRecencyQuery("modified", baseQuery)
	return recency, nil
}

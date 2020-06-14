package main

import (
	"os"
	"unicode"

	"github.com/CGamesPlay/pilikino/pkg/pilikino"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

func buildIndex() (*pilikino.Index, error) {
	if directory != "" {
		if err := os.Chdir(directory); err != nil {
			return nil, err
		}
	}
	index, err := pilikino.NewMemOnlyIndex()
	if err != nil {
		return nil, err
	}
	if err := index.Reindex(); err != nil {
		return nil, err
	}
	return index, nil
}

func parseQuery(queryString string) (query.Query, error) {
	if len(queryString) == 0 {
		return query.NewMatchAllQuery(), nil
	} else if runes := []rune(queryString); unicode.IsLetter(runes[len(runes)-1]) {
		queryString += "*"
	}
	parsed, err := bleve.NewQueryStringQuery(queryString).Parse()
	if err != nil {
		return nil, err
	}
	boolQuery := parsed.(*query.BooleanQuery)
	boolQuery.AddShould(query.NewMatchAllQuery())
	return parsed, nil
}

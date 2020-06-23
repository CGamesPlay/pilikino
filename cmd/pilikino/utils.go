package main

import (
	"os"

	"github.com/CGamesPlay/pilikino/pkg/pilikino"
	"github.com/CGamesPlay/pilikino/pkg/search"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

const (
	// ExitStatusSuccess when the program exits normally.
	ExitStatusSuccess = 0
	// ExitStatusNoResults when the program exits after failing to find any
	// results.
	ExitStatusNoResults = 1
	// ExitStatusError when the program exits after encountering an error.
	ExitStatusError = 2
	// ExitStatusAborted when the program exits because the user aborted the
	// search.
	ExitStatusAborted = 130
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
	var baseQuery query.Query
	if len(queryString) == 0 {
		baseQuery = defaultMatch
	} else {
		parsed, err := search.ParseQuery(queryString)
		if err != nil {
			return nil, err
		}
		baseQuery = query.NewBooleanQuery(nil, []query.Query{parsed, defaultMatch}, nil)
	}
	recency := search.NewRecencyQuery("modified", baseQuery)
	return recency, nil
}

func performSearch(index *pilikino.Index, query query.Query, numResults int) (*bleve.SearchResult, error) {
	sr := bleve.NewSearchRequestOptions(query, numResults, 0, false)
	sr.Fields = []string{"*"}
	sr.Highlight = bleve.NewHighlight()
	sr.Highlight.AddField("content")
	return index.Bleve.Search(sr)
}

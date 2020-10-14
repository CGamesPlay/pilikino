package main

import (
	"fmt"
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

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(ExitStatusError)
	}
}

func getIndex() (*pilikino.Index, error) {
	if err := setupDir(); err != nil {
		return nil, err
	}
	index, err := pilikino.NewMemOnlyIndex()
	if err != nil {
		return nil, err
	}
	return index, nil
}

func parseQuery(queryString string, interactive bool) (query.Query, error) {
	var defaultMatch query.Query
	if interactive && len(queryString) == 0 {
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
		if interactive && queryString[len(queryString)-1] != ' ' {
			// During an interactive search, we treat the final word as a
			// prefix match to enable "co" to match "coffee" as the user types.
			queryString = queryString + "*"
		}
		parsed, err := search.ParseQuery(queryString)
		if err != nil {
			return nil, err
		}
		baseQuery = query.NewBooleanQuery(nil, []query.Query{parsed, defaultMatch}, nil)
	}
	recency := search.NewRecencyQuery("modified", baseQuery)
	recency.SetBoost(0.1)
	return recency, nil
}

func performSearch(index *pilikino.Index, query query.Query, numResults int) (*bleve.SearchResult, error) {
	sr := bleve.NewSearchRequestOptions(query, numResults, 0, false)
	sr.Fields = []string{"*"}
	sr.Highlight = bleve.NewHighlight()
	sr.Highlight.AddField("content")
	return index.Bleve.Search(sr)
}

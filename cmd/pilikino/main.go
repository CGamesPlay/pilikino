package main

import (
	"fmt"
	"os"
	"unicode"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

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

func searcher(index bleve.Index) func(query string, num int) ([]ListItem, error) {
	return func(queryString string, numResults int) ([]ListItem, error) {
		query, err := parseQuery(queryString)
		if err != nil {
			return nil, err
		}
		search := bleve.NewSearchRequestOptions(query, numResults, 0, false)
		search.Highlight = bleve.NewHighlight()
		res, err := index.Search(search)
		if err != nil {
			return nil, err
		}
		results := make([]ListItem, len(res.Hits))
		for i, hit := range res.Hits {
			label := hit.ID
			label += fmt.Sprintf(":%0.4f", hit.Score)
			/*if fragments, ok := hit.Fragments["content"]; ok {
				for _, fragment := range fragments {
					label += strings.Replace(fragment, "\n", " ", -1)
				}
			}*/
			results[i] = ListItem{
				ID:    hit.ID,
				Label: label,
				Score: float32(hit.Score),
			}
		}
		return results, nil
	}
}

func main() {
	var result ListItem

	index, err := createIndex()
	if err != nil {
		goto finish
	}
	if err := indexNotes(index); err != nil {
		goto finish
	}

	result, err = RunInteractive(searcher(index))
	if err != nil {
		goto finish
	}
	fmt.Printf("%v\n", result.ID)

finish:
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

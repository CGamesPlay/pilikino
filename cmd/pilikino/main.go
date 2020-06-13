package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/rivo/tview"
)

var jankyHighlighter = strings.NewReplacer(
	"<mark>", "[#87af87]",
	"</mark>", "[-]",
)

type bleveResult struct {
	hit *search.DocumentMatch
}

func (br *bleveResult) Label() string {
	label := br.hit.ID
	label += fmt.Sprintf(":%0.4f", br.hit.Score)
	return label
}

func (br *bleveResult) Preview(preview *tview.TextView) {
	content := strings.Builder{}
	if fragments, ok := br.hit.Fragments["content"]; ok {
		for _, fragment := range fragments {
			content.WriteString(jankyHighlighter.Replace(tview.Escape(fragment)))
			content.WriteString("\n---\n")
		}
	} else if docContent, ok := br.hit.Fields["content"]; ok {
		content.WriteString(docContent.(string))
	} else {
		for field, value := range br.hit.Fields {
			content.WriteString(fmt.Sprintf("%s:%v\n", field, value))
		}
	}
	preview.SetText(content.String()).SetWordWrap(true).SetDynamicColors(true)
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

func searcher(index bleve.Index) func(query string, num int) (SearchResults, error) {
	return func(queryString string, numResults int) (SearchResults, error) {
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
		results := make(SearchResults, len(res.Hits))
		for i, hit := range res.Hits {
			results[i] = &bleveResult{hit: hit}
		}
		return results, nil
	}
}

func main() {
	var result SearchResult

	index, err := createIndex()
	if err != nil {
		goto finish
	}
	if err := indexNotes(index); err != nil {
		goto finish
	}

	result, err = RunInteractive(searcher(index), true)
	if err != nil {
		goto finish
	} else if result != nil {
		fmt.Printf("%v\n", result.(*bleveResult).hit.ID)
	}

finish:
	// These exit codes are similar to fzf's.
	if err == ErrSearchAborted {
		os.Exit(130)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	} else if result == nil {
		// User selected no match
		os.Exit(1)
	}
}

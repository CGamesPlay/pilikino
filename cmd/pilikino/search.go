package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/CGamesPlay/pilikino/pkg/pilikino"
	"github.com/CGamesPlay/pilikino/pkg/tui"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Interactive search for notes",
	Long:  `Opens a terminal UI featuring an interactive query and displays matching notes as you type.`,
	Run: func(cmd *cobra.Command, args []string) {
		var result tui.SearchResult
		index, err := buildIndex()
		if err != nil {
			goto finish
		}

		result, err = tui.RunInteractive(searcher(index), true)
		if err != nil {
			goto finish
		} else if result != nil {
			fmt.Printf("%v\n", result.(*bleveResult).ID)
		}

	finish:
		// These exit codes are similar to fzf's.
		if err == tui.ErrSearchAborted {
			os.Exit(130)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		} else if result == nil {
			// User selected no match
			os.Exit(1)
		}
	},
}

var jankyHighlighter = strings.NewReplacer(
	"<mark>", "[#87af87]",
	"</mark>", "[-]",
)

type bleveResult struct {
	search.DocumentMatch
}

func (hit *bleveResult) Label() string {
	label := hit.ID
	label += fmt.Sprintf(":%0.4f", hit.Score)
	return label
}

func (hit *bleveResult) Preview(preview *tview.TextView) {
	content := strings.Builder{}
	if fragments, ok := hit.Fragments["content"]; ok {
		for _, fragment := range fragments {
			content.WriteString(jankyHighlighter.Replace(tview.Escape(fragment)))
			content.WriteString("\n---\n")
		}
	} else if docContent, ok := hit.Fields["content"]; ok {
		content.WriteString(docContent.(string))
	} else {
		for field, value := range hit.Fields {
			content.WriteString(fmt.Sprintf("%s:%v\n", field, value))
		}
	}
	preview.SetText(content.String()).SetWordWrap(true).SetDynamicColors(true)
}

func searcher(index *pilikino.Index) func(query string, num int) (tui.SearchResults, error) {
	return func(queryString string, numResults int) (tui.SearchResults, error) {
		query, err := parseQuery(queryString)
		if err != nil {
			return nil, err
		}
		search := bleve.NewSearchRequestOptions(query, numResults, 0, false)
		search.Highlight = bleve.NewHighlight()
		res, err := index.Bleve.Search(search)
		if err != nil {
			return nil, err
		}
		results := make(tui.SearchResults, len(res.Hits))
		for i, hit := range res.Hits {
			results[i] = &bleveResult{*hit}
		}
		return results, nil
	}
}

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

var searchKeys []string

func init() {
	searchCmd.Flags().StringSliceVar(&searchKeys, "expect", []string{}, "list of keys to accept a result")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Interactive search for notes",
	Long: `Opens a terminal UI featuring an interactive query and displays matching notes as you type.

Using the --expect flag, you can build integrations with other commands. If this option is set, the first line of output for a successful search will be the name of the key that was typed to accept the search. Example key names: f1, ctrl-v, enter, alt-shift-s. Note that enter will always accept the search, but will not be printed unless --expect is specified. You can use --expect=enter to make this explicit.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			index    *pilikino.Index
			result   *tui.InteractiveResults
			t        *tui.Tui
			asyncErr chan error
			err      error
		)

		expectedKeys := make([]tui.Key, len(searchKeys)+1)
		for i, name := range searchKeys {
			var key tui.Key
			key, err = tui.ParseKey(name)
			if err != nil {
				goto fail
			}
			expectedKeys[i] = key
		}
		expectedKeys[len(searchKeys)] = tui.KeyEnter

		if index, err = getIndex(); err != nil {
			goto fail
		}

		t = tui.NewTui(searcher(index), true)
		t.ExpectedKeys = expectedKeys
		result, err = t.Run(func() {
			asyncErr = make(chan error, 1)
			go indexAsync(index, t, asyncErr)
		})
		// If an async error happened, take priority over the ErrSearchAborted.
		select {
		case err = <-asyncErr:
		default:
		}
		if err != nil {
			goto fail
		}
		if len(expectedKeys) > 1 {
			if result.Action < len(searchKeys) {
				fmt.Printf("%s\n", searchKeys[result.Action])
			} else {
				fmt.Println("enter")
			}
		}
		for _, hit := range result.Results {
			fmt.Printf("%v\n", hit.(*bleveResult).ID)
		}

	fail:
		// These exit codes are similar to fzf's.
		if err == tui.ErrSearchAborted {
			os.Exit(130)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		} else if len(result.Results) == 0 {
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
	label += fmt.Sprintf(":%.4f", hit.Score)
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
		content.WriteString(tview.Escape(docContent.(string)))
	}
	preview.SetScrollable(true).SetText(content.String()).SetWordWrap(true).SetDynamicColors(true).ScrollTo(0, 0)
}

func searcher(index *pilikino.Index) func(query string, num int) (tui.SearchResult, error) {
	return func(queryString string, numResults int) (tui.SearchResult, error) {
		query, err := parseQuery(queryString, true)
		if err != nil {
			sr := tui.SearchResult{QueryError: err}
			return sr, nil
		}
		sr := bleve.NewSearchRequestOptions(query, numResults, 0, false)
		sr.Fields = []string{"content"}
		sr.Highlight = bleve.NewHighlight()
		bleveRes, err := index.Bleve.Search(sr)
		if err != nil {
			return tui.SearchResult{}, err
		}
		hits := make([]tui.Document, len(bleveRes.Hits))
		for i, hit := range bleveRes.Hits {
			hits[i] = &bleveResult{*hit}
		}
		res := tui.SearchResult{
			Results: hits,
			// This works because we have a built-in matchAll for all queries
			TotalCandidates: bleveRes.Total,
		}
		return res, nil
	}
}

func indexAsync(index *pilikino.Index, t *tui.Tui, asyncErr chan error) {
	progress := func(p pilikino.IndexProgress) {
		t.SetStatusText(fmt.Sprintf("Scanned %d files", p.Scanned))
		t.Refresh()
	}
	if err := index.Reindex(progress); err != nil {
		asyncErr <- err
		t.Stop()
	}
}

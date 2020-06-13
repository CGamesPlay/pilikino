package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/CGamesPlay/pilikino/pkg/tui"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
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

		index, err := createIndex()
		if err != nil {
			goto finish
		}
		if err := indexNotes(index); err != nil {
			goto finish
		}

		result, err = tui.RunInteractive(searcher(index), true)
		if err != nil {
			goto finish
		} else if result != nil {
			fmt.Printf("%v\n", result.(*bleveResult).hit.ID)
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

func searcher(index bleve.Index) func(query string, num int) (tui.SearchResults, error) {
	return func(queryString string, numResults int) (tui.SearchResults, error) {
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
		results := make(tui.SearchResults, len(res.Hits))
		for i, hit := range res.Hits {
			results[i] = &bleveResult{hit: hit}
		}
		return results, nil
	}
}

func createIndexMapping() (mapping.IndexMapping, error) {
	// a generic reusable mapping for english text
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	// a generic reusable mapping for keyword text
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	noteMapping := bleve.NewDocumentMapping()
	noteMapping.AddFieldMappingsAt("Filename", keywordFieldMapping)
	noteMapping.AddFieldMappingsAt("Content", englishTextFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("note", noteMapping)
	indexMapping.DefaultType = "note"
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping, nil
}

func createIndex() (bleve.Index, error) {
	mapping, err := createIndexMapping()
	if err != nil {
		return nil, err
	}
	return bleve.NewMemOnly(mapping)
}

type noteData struct {
	Filename string    `json:"filename"`
	ModTime  time.Time `json:"mtime"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
}

func indexNote(batch *bleve.Batch, path string, info os.FileInfo) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	note := noteData{
		Filename: path,
		ModTime:  info.ModTime(),
		Title:    path,
		Content:  string(content),
	}
	return batch.Index(path, note)
}

func indexNotes(index bleve.Index) error {
	batch := index.NewBatch()
	root := "/Users/rpatterson/Seafile/Notes"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".md") {
			return indexNote(batch, path, info)
		}
		return nil
	})
	if err != nil {
		return err
	}
	index.Batch(batch)

	return nil
}

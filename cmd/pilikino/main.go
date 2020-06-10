package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/highlight/highlighter/ansi"
)

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

type Note struct {
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
	note := Note{
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

func main() {
	index, err := createIndex()
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := indexNotes(index); err != nil {
		fmt.Println(err)
		return
	}

	// search for some text
	query := bleve.NewMatchQuery(os.Args[1])
	search := bleve.NewSearchRequest(query)
	search.Highlight = bleve.NewHighlightWithStyle(ansi.Name)
	res, err := index.Search(search)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v matches, showing %v through %v, took %v\n", res.Total, res.Request.From+1, len(res.Hits)+res.Request.From, res.Took)
	for i, hit := range res.Hits {
		fmt.Printf("%2d. %s (%f)\n", i+res.Request.From+1, hit.ID, hit.Score)
		if fragments, ok := hit.Fragments["content"]; ok {
			for _, fragment := range fragments {
				fmt.Printf("\t%s\n", strings.Replace(fragment, "\n", " ", -1))
			}
		}
	}
}

package pilikino

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
)

// Index is the top-level data structure used to search notes.
type Index struct {
	// Bleve is the actual Bleve index used by this Index.
	Bleve bleve.Index
}

// NewMemOnlyIndex creates a new pilikino index that will not be persisted to
// disk.
func NewMemOnlyIndex() (*Index, error) {
	mapping, err := createIndexMapping()
	if err != nil {
		return nil, err
	}
	bleve, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}
	index := &Index{
		Bleve: bleve,
	}
	return index, nil
}

func createIndexMapping() (mapping.IndexMapping, error) {
	// a generic reusable mapping for english text
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName
	englishTextFieldMapping.Store = true

	// a generic reusable mapping for keyword text
	simpleFieldMapping := bleve.NewTextFieldMapping()
	simpleFieldMapping.Analyzer = simple.Name
	simpleFieldMapping.Store = true

	noteMapping := bleve.NewDocumentMapping()
	noteMapping.AddFieldMappingsAt("Filename", simpleFieldMapping)
	noteMapping.AddFieldMappingsAt("Content", englishTextFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("note", noteMapping)
	indexMapping.DefaultType = "note"
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping, nil
}

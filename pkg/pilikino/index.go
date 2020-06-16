package pilikino

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/analysis/datetime/optional"
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

// DocumentCount returns the total number of documents in the index
func (index *Index) DocumentCount() (uint64, error) {
	q := bleve.NewMatchAllQuery()
	search := bleve.NewSearchRequestOptions(q, 0, 0, false)
	res, err := index.Bleve.Search(search)
	if err != nil {
		return 0, err
	}
	return res.Total, nil
}

func createIndexMapping() (mapping.IndexMapping, error) {
	// reusable mapping for english text
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName
	englishTextFieldMapping.Store = true

	// reusable mapping for unstemmed words
	simpleFieldMapping := bleve.NewTextFieldMapping()
	simpleFieldMapping.Analyzer = simple.Name
	simpleFieldMapping.Store = true

	// reusable mapping for dates
	dateFieldMapping := bleve.NewDateTimeFieldMapping()
	dateFieldMapping.DateFormat = optional.Name

	noteMapping := bleve.NewDocumentMapping()
	noteMapping.AddFieldMappingsAt("Filename", simpleFieldMapping)
	noteMapping.AddFieldMappingsAt("Title", simpleFieldMapping)
	noteMapping.AddFieldMappingsAt("Content", englishTextFieldMapping)
	noteMapping.AddFieldMappingsAt("Tags", simpleFieldMapping)
	noteMapping.AddFieldMappingsAt("ModTime", dateFieldMapping)
	noteMapping.AddFieldMappingsAt("CreateTime", dateFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("note", noteMapping)
	indexMapping.DefaultType = "note"
	indexMapping.DefaultAnalyzer = en.AnalyzerName

	return indexMapping, nil
}

package pilikino

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
)

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

// Reindex crawls all files in the root directory and adds them to the index.
// TODO: It does not remove deleted files from the index.
func (index *Index) Reindex() error {
	batch := index.Bleve.NewBatch()
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
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
	index.Bleve.Batch(batch)

	return nil
}

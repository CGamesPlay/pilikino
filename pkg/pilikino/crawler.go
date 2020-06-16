package pilikino

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
)

const progressInterval = 200 * time.Millisecond

// IndexProgress holds data relating to an in-progress scan of the database.
type IndexProgress struct {
	Scanned uint64
	Total   uint64
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

// Reindex crawls all files in the root directory and adds them to the index.
// The progress function, if provided, will be called periodically during the
// indexing. It is guaranteed to be called at least once, and the last call
// will be just before Reindex returns.
// TODO: It does not remove deleted files from the index.
func (index *Index) Reindex(progressFunc func(IndexProgress)) error {
	progress := IndexProgress{}
	batchTime := time.Now().Add(progressInterval)
	batch := index.Bleve.NewBatch()
	cycleBatch := func() {
		batchTime = time.Now().Add(progressInterval)
		index.Bleve.Batch(batch)
		batch = index.Bleve.NewBatch()
		if progressFunc != nil {
			progressFunc(progress)
		}
	}
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		progress.Scanned++
		progress.Total++
		if strings.HasSuffix(path, ".md") {
			err = indexNote(batch, path, info)
		}
		if time.Now().After(batchTime) {
			cycleBatch()
		}
		return err
	})
	if err != nil {
		return err
	}
	cycleBatch()

	return nil
}

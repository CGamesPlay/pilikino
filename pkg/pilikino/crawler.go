package pilikino

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const progressInterval = 200 * time.Millisecond

// IndexProgress holds data relating to an in-progress scan of the database.
type IndexProgress struct {
	Scanned uint64
	Total   uint64
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
		err = indexNote(index, batch, path, info)
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

var globEscaper = strings.NewReplacer(
	"*", "\\*",
	"?", "\\?",
	"[", "\\[",
	"\\", "\\\\",
)

// resolveLink will attempt to locate a note located at `to` relative to
// `from`. This may resolve links that are specified from partial filenames.
// This method will only return an error if there are multiple ambiguous files
// that could be referred to. If there are no possible files, it will return an
// empty string and nil error.
func (index *Index) resolveLink(from, to string) (string, error) {
	fromDir := filepath.Dir(from)
	toDir, toBase := filepath.Split(to)
	toGlob := fmt.Sprintf("*%v*", globEscaper.Replace(toBase))
	glob := filepath.Join(".", globEscaper.Replace(fromDir), globEscaper.Replace(toDir), toGlob)
	matches, err := filepath.Glob(glob)
	if err != nil {
		panic("bad pattern")
	} else if len(matches) > 1 {
		return "", errors.New("ambiguous link")
	} else if len(matches) == 0 {
		return "", nil
	}
	return matches[0], nil
}

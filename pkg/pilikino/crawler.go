package pilikino

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrAmbiguousLink indicates that a relative link could have resolved to
// multiple files.
var ErrAmbiguousLink = errors.New("ambiguous link")
// ErrDeadLink indicates that a relative link does not resolve to an existing
// file.
var ErrDeadLink = errors.New("dead link")

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
	`*`, `\*`,
	`?`, `\?`,
	`[`, `\[`,
	`\`, `\\`,
)

// getRelativePath returns the pathname made relative to the cwd. It will fail
// if the path points outside of the cwd.
func getRelativePath(dir string) (string, error) {
	dir = filepath.Clean(dir)
	if filepath.IsAbs(dir) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir, err = filepath.Rel(cwd, dir)
		if err != nil {
			return "", err
		}
	}
	if filepath.HasPrefix(dir, filepath.FromSlash("../")) {
		return "", errors.New("invalid relative link")
	}
	return dir, nil
}

// ResolveLink will attempt to locate a note located at `to` relative to
// `from`. This may resolve links that are specified from partial filenames. If
// an error is not returned, the path will always be relative to the cwd and
// not begin with `..`.
func (index *Index) ResolveLink(from, to string) (string, error) {
	// Step 1: convert the link to relative to cwd (or absolute)
	fromDir := filepath.Dir(from)
	if !filepath.IsAbs(to) {
		to = filepath.Join(fromDir, to)
	}
	// Step 2: validate directory
	toDir, toBase := filepath.Split(to)
	toDir, err := getRelativePath(toDir)
	if err != nil {
		return "", err
	}
	// Step 3: resolve partial filenames
	toGlob := fmt.Sprintf("*%v*", globEscaper.Replace(toBase))
	glob := filepath.Join(globEscaper.Replace(toDir), toGlob)
	matches, err := filepath.Glob(glob)
	if err != nil {
		panic("bad pattern")
	} else if len(matches) > 1 {
		return "", ErrAmbiguousLink
	} else if len(matches) == 0 {
		return "", ErrDeadLink
	}
	return matches[0], nil
}

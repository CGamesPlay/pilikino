package file

import (
	"net/url"
	"strings"

	fs "github.com/relab/wrfs"

	"github.com/CGamesPlay/pilikino/lib/notedb"
)

func init() {
	notedb.RegisterFormat(notedb.FormatDescription{
		ID:            "file",
		Description:   "Directory of files",
		Documentation: `This file format corresponds to a directory of Markdown files.`,
		Open:          OpenDatabase,
		Detect:        Detect,
	})
}

type Database struct {
	fs.FS
}

// OpenDatabase is the entrypoint for the file format.
func OpenDatabase(dbURL *url.URL) (notedb.Database, error) {
	return &Database{fs.DirFS(dbURL.Path)}, nil
}

// Detect determines if the URL is likely to be a note database.
func Detect(dbURL *url.URL) notedb.DetectResult {
	return notedb.DetectResultUnknown
}

func (db *Database) Open(path string) (fs.File, error) {
	f, err := db.FS.Open(path)
	if err != nil {
		return nil, err
	}
	return &file{f}, nil
}

type file struct {
	fs.File
}

func (f *file) ReadDir(n int) ([]fs.DirEntry, error) {
	rdf, ok := f.File.(fs.ReadDirFile)
	if !ok {
		return nil, fs.ErrInvalid
	}
	entries, err := rdf.ReadDir(n)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		info, err := entries[i].Info()
		if err != nil {
			return nil, err
		}
		entries[i] = &fileInfo{info}
	}
	return entries, nil
}

type fileInfo struct {
	fs.FileInfo
}

var _ notedb.FileInfo = (*fileInfo)(nil)

func (i *fileInfo) Type() fs.FileMode          { return i.Mode().Type() }
func (i *fileInfo) Info() (fs.FileInfo, error) { return i, nil }
func (i *fileInfo) IsNote() bool               { return strings.HasSuffix(i.Name(), ".md") }

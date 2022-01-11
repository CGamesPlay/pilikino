package file

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	fs "github.com/relab/wrfs"

	"github.com/CGamesPlay/pilikino/lib/markdown/parser"
	"github.com/CGamesPlay/pilikino/lib/notedb"
	"github.com/yuin/goldmark/ast"
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

var _ fs.MkdirAllFS = (*Database)(nil)
var _ fs.OpenFileFS = (*Database)(nil)
var _ fs.ChtimesFS = (*Database)(nil)

// OpenDatabase is the entrypoint for the file format.
func OpenDatabase(dbURL *url.URL) (notedb.Database, error) {
	return &Database{fs.DirFS(dbURL.Path)}, nil
}

// Detect determines if the URL is likely to be a note database.
func Detect(dbURL *url.URL) notedb.DetectResult {
	return notedb.DetectResultUnknown
}

func (db *Database) Open(path string) (fs.File, error) {
	return db.OpenFile(path, os.O_RDONLY, 0)
}

func (db *Database) OpenFile(path string, flag int, perm fs.FileMode) (fs.File, error) {
	f, err := fs.OpenFile(db.FS, path, flag, perm)
	if err != nil {
		return nil, err
	}
	return &file{f, nil}, nil
}

func (db *Database) MkdirAll(path string, perm fs.FileMode) error {
	return fs.MkdirAll(db.FS, path, perm)
}

func (db *Database) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return fs.Chtimes(db.FS, name, atime, mtime)
}

type file struct {
	fs.File
	data []byte
}

var _ notedb.Note = (*file)(nil)
var _ fs.ReadDirFile = (*file)(nil)
var _ fs.WriteFile = (*file)(nil)

func (f *file) Stat() (fs.FileInfo, error) {
	info, err := f.File.Stat()
	return &fileInfo{info}, err
}

func (f *file) IsNote() bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.(*fileInfo).IsNote()
}

func (f *file) ParseAST() (ast.Node, error) {
	doc, err := parser.Parse(f.Data())
	return doc, err
}

func (f *file) Data() []byte {
	if f.data == nil {
		var err error
		f.data, err = io.ReadAll(f.File)
		if err != nil {
			panic(fmt.Errorf("failed to read note: %w", err))
		}
		fmt.Printf("Read data %#v\n", f.data)
	}
	return f.data
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

func (f *file) Write(p []byte) (n int, err error) {
	return fs.Write(f.File, p)
}

type fileInfo struct {
	fs.FileInfo
}

var _ notedb.NoteInfo = (*fileInfo)(nil)

func (i *fileInfo) Type() fs.FileMode          { return i.Mode().Type() }
func (i *fileInfo) Info() (fs.FileInfo, error) { return i, nil }
func (i *fileInfo) IsNote() bool               { return strings.HasSuffix(i.Name(), ".md") }

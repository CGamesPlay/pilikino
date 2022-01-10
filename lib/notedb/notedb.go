package notedb

import (
	"fmt"
	"io/fs"
	"net/url"
	"strings"
)

// Database provides methods to access a database of notes. Implementations are
// required to be goroutine-safe. The canonical format of a note database is a
// directory structure, however all fs.Files must satisfy the notedb.File
// interface, and all fs.FileInfos must satisfy the notedb.FileInfo interface.
type Database fs.FS

// OpenDatabase loads the database at the given URL using the appropriate
// format. The format is determined by looking at the scheme of the provided
// URL up to the first `+`, therefore a URL like "evernote+https://..." would
// use the "evernote" loader.
func OpenDatabase(dbURL *url.URL) (Database, error) {
	scheme := dbURL.Scheme
	if idx := strings.IndexByte(scheme, '+'); idx != -1 {
		scheme = scheme[0:idx]
	}
	format, ok := registeredFormats[scheme]
	if !ok {
		return nil, fmt.Errorf("%s is unrecognized", scheme)
	}
	return format.Open(dbURL)
}

// File provides raw access to database contents, but also provides a
// high-level methods specific to notes.
type File interface {
	fs.File
	ExtraMethod()
}

type FileInfo interface {
	fs.FileInfo
	ExtraMethod()
}

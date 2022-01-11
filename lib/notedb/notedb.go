package notedb

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/CGamesPlay/pilikino/lib/markdown/renderer"
	fs "github.com/relab/wrfs"
	"github.com/yuin/goldmark/ast"
)

// Database provides methods to access a database of notes. Implementations are
// required to be goroutine-safe. The canonical format of a note database is a
// directory structure, however all fs.Files must satisfy the notedb.File
// interface, and all fs.FileInfos must satisfy the notedb.FileInfo interface.
type Database = fs.FS

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

// Note extends fs.File with additional methods specific to note databases.
type Note interface {
	fs.File
	IsNote() bool
	// Parses the Note data into an ast.Node and returns it along with any
	// errors that occurred. Note that an AST is always returned if possible,
	// and any errors returned in this case are non-fatal.
	ParseAST() (ast.Node, error)
	// Return the raw data of the note as a slice. This provides copy-free
	// access to the underlying data read from the database.
	Data() []byte
}

// WriteASTNote extends Note with additional methods to rewrite Markdown during
// the writing process.
type WriteASTNote interface {
	Note
	WriteAST(ast.Node) error
}

func WriteAST(n Note, node ast.Node, data []byte) error {
	if wn, ok := n.(WriteASTNote); ok {
		return wn.WriteAST(node)
	} else if wf, ok := n.(fs.WriteFile); ok {
		r := renderer.NewRenderer()
		return r.Render(wf, data, node)
	}
	return fs.ErrUnsupported
}

// NoteInfo extends fs.FileInfo with additional methods specific to note
// databases.
type NoteInfo interface {
	fs.FileInfo
	IsNote() bool
}

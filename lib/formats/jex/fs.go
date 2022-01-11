package jex

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/CGamesPlay/pilikino/lib/markdown/parser"
	"github.com/CGamesPlay/pilikino/lib/notedb"
	fs "github.com/relab/wrfs"
	"github.com/yuin/goldmark/ast"
	"go.uber.org/multierr"
)

var resourcesFolder = "_resources"

var escapePathReplacer = strings.NewReplacer(
	"<", "(",
	">", ")",
	":", "",
	"\"", "",
	"/", "",
	"\\", "",
	"|", "",
	"?", "",
	"*", "",
	"\n", " ",
	"\t", " ",
)

func escapePath(in string) string {
	return strings.Join(strings.Fields(escapePathReplacer.Replace(in)), " ")
}

func genName(obj *jexObject, long bool) string {
	title := escapePath(obj.Title)
	if long {
		title = title + "-" + obj.ID
	}
	if obj.Type == TypeNote {
		return title + ".md"
	}
	return title
}

func markName(name string, list map[string]int) {
	count, _ := list[name]
	list[name] = count + 1
}

func deduplicateNames(items []*jfsEntry, usedNames map[string]int) {
	for _, entry := range items {
		if count, _ := usedNames[entry.name]; count <= 1 {
			continue
		}
		if entry.object != nil {
			entry.name = genName(entry.object, true)
		}
	}
}

func registerPaths(items []*jfsEntry, parentPath string, lookup map[string]string) {
	for _, entry := range items {
		if entry.object != nil {
			lookup[entry.object.ID] = parentPath + "/" + entry.name
		}
	}
}

type JoplinFS struct {
	root       *jfsEntry
	pathLookup map[string]string
}

func newJoplinFS(jex *JEX) (*JoplinFS, error) {
	ret := &JoplinFS{
		&jfsEntry{nil, "", []*jfsEntry{}},
		map[string]string{},
	}
	itemsByParent := map[string][]*jexObject{}
	for _, child := range jex.objects {
		parentItems, _ := itemsByParent[child.ParentID]
		parentItems = append(parentItems, child)
		itemsByParent[child.ParentID] = parentItems
	}

	queue := make([]*jfsEntry, 1)
	queue[0] = ret.root
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]

		// Populate all items, names are not finalized.
		var parentID string
		var parentPath string
		if parent.object != nil {
			parentID = parent.object.ID
			parentPath = ret.pathLookup[parentID]
		}
		items, _ := itemsByParent[parentID]
		usedNames := map[string]int{}
		hasResources := false
		for _, item := range items {
			if item.Type == TypeResource {
				hasResources = true
				continue
			} else if item.Type != TypeFolder && item.Type != TypeNote {
				continue
			}

			var contents []*jfsEntry
			if item.Type == TypeFolder {
				contents = make([]*jfsEntry, 0)
			}

			entry := &jfsEntry{
				item,
				genName(item, false),
				contents,
			}
			parent.items = append(parent.items, entry)
			markName(entry.name, usedNames)

			if contents != nil {
				queue = append(queue, entry)
			}
		}
		if hasResources {
			resources := &jfsEntry{
				nil,
				resourcesFolder,
				make([]*jfsEntry, 0),
			}
			parent.items = append(parent.items, resources)
			markName(resourcesFolder, usedNames)

			resourceNames := map[string]int{}
			for _, item := range items {
				if item.Type != TypeResource {
					continue
				}

				entry := &jfsEntry{
					item,
					genName(item, false),
					nil,
				}
				resources.items = append(resources.items, entry)
				markName(entry.name, resourceNames)
			}
			deduplicateNames(resources.items, resourceNames)
			registerPaths(resources.items, parentPath+"/"+resourcesFolder, ret.pathLookup)
		}

		deduplicateNames(parent.items, usedNames)
		registerPaths(parent.items, parentPath, ret.pathLookup)
	}

	return ret, nil
}

// Open satisfies notedb.Database.
func (j *JoplinFS) Open(path string) (fs.File, error) {
	if !fs.ValidPath(path) {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrInvalid}
	}
	components := strings.Split(path, "/")
	if components[0] == "." {
		components = components[1:]
	}
	entry, err := j.root.getEntry(components)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: path, Err: err}
	}
	return &jfsHandle{entry, j, 0}, nil
}

type jfsEntry struct {
	object *jexObject
	name   string
	items  []*jfsEntry
}

func (j *jfsEntry) getEntry(components []string) (*jfsEntry, error) {
	if len(components) == 0 {
		return j, nil
	}

	if j.items == nil {
		return nil, fs.ErrInvalid
	}

	for _, item := range j.items {
		if item.name == components[0] {
			return item.getEntry(components[1:])
		}
	}

	return nil, fs.ErrNotExist
}

type jfsHandle struct {
	*jfsEntry
	fs     *JoplinFS
	cursor int
}

var _ fs.ReadDirFile = (*jfsHandle)(nil)
var _ notedb.Note = (*jfsHandle)(nil)

func (j *jfsHandle) Stat() (fs.FileInfo, error) {
	var modTime time.Time
	if j.object != nil {
		modTime = j.object.ModTime
	}
	return &jfsNoteInfo{
		j.name,
		0,
		fs.ModeDir | 0555,
		modTime,
		j.object != nil && j.object.Type == TypeNote,
	}, nil
}

func (j *jfsHandle) Read(ret []byte) (count int, err error) {
	if j.object == nil || j.object.Data == nil {
		return 0, &fs.PathError{Op: "read", Path: j.name, Err: fs.ErrInvalid}
	}
	if j.cursor == len(j.object.Data) {
		return 0, io.EOF
	}
	count = len(ret)
	if count+j.cursor >= len(j.object.Data) {
		count = len(j.object.Data) - j.cursor
	}
	copy(ret[0:count], j.object.Data[j.cursor:j.cursor+count])
	j.cursor += count
	return
}

func (j *jfsHandle) Close() error {
	return nil
}

func (j *jfsHandle) ReadDir(n int) ([]fs.DirEntry, error) {
	if n != -1 {
		panic("partial directory reads are not supported")
	}

	if j.items == nil {
		return nil, fs.ErrInvalid
	}

	ret := make([]fs.DirEntry, len(j.items))
	for i, item := range j.items {
		var mode fs.FileMode
		if item.object == nil || item.object.Type == TypeFolder {
			mode = fs.ModeDir | 0555
		} else {
			mode = 0444
		}
		size := 0
		var modTime time.Time
		if item.object != nil {
			size = len(item.object.Data)
			modTime = item.object.ModTime
		}
		ret[i] = &jfsNoteInfo{
			item.name,
			int64(size),
			mode,
			modTime,
			item.object != nil && item.object.Type == TypeNote,
		}
	}
	return ret, nil
}

func (j *jfsHandle) IsNote() bool {
	return j.object != nil && j.object.Type == TypeNote
}

func (j *jfsHandle) ParseAST() (ast.Node, error) {
	if j.object == nil || j.object.Type != TypeNote {
		return nil, fs.ErrInvalid
	}
	base, ok := j.fs.pathLookup[j.object.ID]
	if !ok {
		return nil, fmt.Errorf("cannot find own ID in path lookup")
	}
	base, _ = filepath.Split(base)
	doc, err := parser.Parse(j.object.Data)

	replaceLink := func(id string, orig []byte) []byte {
		found, ok := j.fs.pathLookup[id]
		if !ok {
			err = multierr.Append(err, fmt.Errorf("dead link: %s", orig))
			return orig
		}

		relative, err := filepath.Rel(base, found)
		if err != nil {
			// Should never happen, since all paths are absolute
			panic(err)
		}
		return []byte(relative)
	}

	linkResolver := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := n.(*ast.Link); ok && entering {
			dest := string(link.Destination)
			if strings.HasPrefix(dest, ":/") {
				link.Destination = replaceLink(dest[2:], link.Destination)
			} else if u, err := url.Parse(dest); err == nil {
				if u.Scheme == "joplin" && u.Host == "x-callback-url" && u.Path == "/openNote" {
					if target, ok := u.Query()["id"]; ok {
						link.Destination = replaceLink(target[0], link.Destination)
					}
				}
			}
		} else if img, ok := n.(*ast.Image); ok && entering {
			dest := string(img.Destination)
			if strings.HasPrefix(dest, ":/") {
				img.Destination = replaceLink(dest[2:], img.Destination)
			}
		}
		return ast.WalkContinue, nil
	}

	if walkErr := ast.Walk(doc, linkResolver); walkErr != nil {
		err = multierr.Append(err, walkErr)
	}
	return doc, err
}

func (j *jfsHandle) Data() []byte {
	if j.object == nil {
		return nil
	}
	return j.object.Data
}

type jfsNoteInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isNote  bool
}

var _ notedb.NoteInfo = (*jfsNoteInfo)(nil)

func (i *jfsNoteInfo) Name() string               { return i.name }
func (i *jfsNoteInfo) Size() int64                { return i.size }
func (i *jfsNoteInfo) Mode() fs.FileMode          { return i.mode }
func (i *jfsNoteInfo) ModTime() time.Time         { return i.modTime }
func (i *jfsNoteInfo) IsDir() bool                { return i.mode&fs.ModeDir != 0 }
func (i *jfsNoteInfo) Sys() interface{}           { return nil }
func (i *jfsNoteInfo) Type() fs.FileMode          { return i.mode.Type() }
func (i *jfsNoteInfo) Info() (fs.FileInfo, error) { return i, nil }
func (i *jfsNoteInfo) IsNote() bool               { return i.isNote }

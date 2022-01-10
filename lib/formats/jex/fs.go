package jex

import (
	"io/fs"
	"strings"
	"time"
)

var resourcesFolder = "_resources"
var timeNotImplemented time.Time

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

type JoplinFS struct {
	root *jfsEntry
}

func newJoplinFS(jex *JEX) (*JoplinFS, error) {
	ret := &JoplinFS{
		&jfsEntry{nil, "", []*jfsEntry{}},
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
		if parent.object != nil {
			parentID = parent.object.ID
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
		}

		deduplicateNames(parent.items, usedNames)
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
	file, err := j.root.open(components)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: path, Err: err}
	}
	return file, nil
}

type jfsEntry struct {
	object *jexObject
	name   string
	items  []*jfsEntry
}

var _ fs.ReadDirFile = (*jfsEntry)(nil)

func (j *jfsEntry) Stat() (fs.FileInfo, error) {
	return &jfsFileInfo{
		j.name,
		0,
		fs.ModeDir | 0555,
		timeNotImplemented,
	}, nil
}

func (j *jfsEntry) Read(ret []byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: "/", Err: fs.ErrInvalid}
}

func (j *jfsEntry) Close() error {
	return nil
}

func (j *jfsEntry) open(components []string) (*jfsEntry, error) {
	if len(components) == 0 {
		return j, nil
	}

	if j.items == nil {
		return nil, fs.ErrInvalid
	}

	for _, item := range j.items {
		if item.name == components[0] {
			return item.open(components[1:])
		}
	}

	return nil, fs.ErrNotExist
}

func (j *jfsEntry) ReadDir(n int) ([]fs.DirEntry, error) {
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
		if item.object != nil {
			size = len(item.object.Data)
		}
		ret[i] = &jfsFileInfo{
			item.name,
			int64(size),
			mode,
			timeNotImplemented,
		}
	}
	return ret, nil
}

type jfsFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func (i *jfsFileInfo) Name() string               { return i.name }
func (i *jfsFileInfo) Size() int64                { return i.size }
func (i *jfsFileInfo) Mode() fs.FileMode          { return i.mode }
func (i *jfsFileInfo) ModTime() time.Time         { return i.modTime }
func (i *jfsFileInfo) IsDir() bool                { return i.mode&fs.ModeDir != 0 }
func (i *jfsFileInfo) Sys() interface{}           { return nil }
func (i *jfsFileInfo) Type() fs.FileMode          { return i.mode.Type() }
func (i *jfsFileInfo) Info() (fs.FileInfo, error) { return i, nil }

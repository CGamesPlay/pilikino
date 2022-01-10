package jex

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/CGamesPlay/pilikino/lib/notedb"
)

const (
	TypeNote = iota + 1
	TypeFolder
	TypeSetting
	TypeResource
	TypeTag
	TypeNoteTag
	TypeSearch
	TypeAlarm
	TypeMasterKey
	TypeItemChange
	TypeNoteResource
	TypeResourceLocalState
	TypeRevision
	TypeMigration
	TypeSmartFilter
	TypeCommand
)

// JEX represents an in-memory copy of a JEX file.
type JEX struct {
	objects []*jexObject
}

func newJEX(file io.Reader) (*JEX, error) {
	// Load the objects and object data into separate structs. To Joplin, an
	// attached file or inline image is a "resource", and the resource data is
	// stored separately in the export. The objectData object is the unparsed
	// Joplin object.
	objectData := make(map[string][]byte)
	resourceData := make(map[string][]byte)
	archive := tar.NewReader(file)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		data, err := ioutil.ReadAll(archive)
		if err != nil {
			return nil, fmt.Errorf("while reading %s: %w", header.Name, err)
		}
		if strings.HasPrefix(header.Name, "resources/") {
			resourceData[header.Name[len("resources/")+1:]] = data
		} else if strings.HasSuffix(header.Name, ".md") {
			objectData[header.Name[0:len(header.Name)-len(".md")]] = data
		} else {
			return nil, fmt.Errorf("unsupported file in JEX: %s", header.Name)
		}
	}

	ret := &JEX{make([]*jexObject, 0, len(objectData))}

	for id, data := range objectData {
		object, err := newjexObject(string(data))
		if err != nil {
			return nil, fmt.Errorf("while loading %s: %w", id, err)
		}
		if object.Type == TypeResource {
			if object.Data != nil {
				return nil, fmt.Errorf("while loading %s: resource has data", id)
			}
			object.Data = resourceData[id]
		}
		ret.objects = append(ret.objects, object)
	}

	return ret, nil
}

type jexObject struct {
	ID       string
	Type     int
	Title    string
	Data     []byte
	ParentID string
}

func newjexObject(rawObject string) (*jexObject, error) {
	ret := &jexObject{}
	props := make(map[string]string)

	// Proceed upwards from the end of the file until a blank line is
	// encountered. This separates the props from the body of the object.
	var lineEnd = len(rawObject)
	for {
		lineStart := strings.LastIndexByte(rawObject[0:lineEnd], '\n')
		if lineStart+1 >= lineEnd {
			// No more props
			lineEnd = lineStart
			break
		}
		line := rawObject[lineStart+1 : lineEnd]
		sep := strings.IndexByte(line, ':')
		if sep == -1 {
			return nil, fmt.Errorf("object properties invalid")
		}
		key := strings.TrimSpace(line[:sep])
		val := strings.TrimSpace(line[sep+1:])
		props[key] = val
		lineEnd = lineStart
		if lineEnd == -1 {
			break
		}
	}

	id, ok := props["id"]
	if !ok {
		return nil, fmt.Errorf("object has no ID")
	}
	ret.ID = id
	typeStr, ok := props["type_"]
	if !ok {
		return nil, fmt.Errorf("object has no type")
	}
	typeInt, err := strconv.Atoi(typeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid type: %w", err)
	}
	ret.Type = typeInt
	ret.ParentID, _ = props["parent_id"]

	if lineEnd != 0 {
		ret.Title = rawObject[0:strings.IndexByte(rawObject, '\n')]
		if len(ret.Title)+1 < lineEnd {
			// Extract the body if present
			ret.Data = []byte(rawObject[len(ret.Title)+2 : lineEnd])
		}
	}

	return ret, nil
}

func init() {
	notedb.RegisterFormat(notedb.FormatDescription{
		ID:            "joplin-export",
		Description:   "Joplin export (JEX)",
		Documentation: `This is the official export format for Joplin.`,
		Open:          OpenDatabase,
		Detect:        Detect,
	})
}

// OpenDatabase is the entrypoint for the Joplin JEX format.
func OpenDatabase(dbURL *url.URL) (notedb.Database, error) {
	file, err := os.Open(dbURL.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	jex, err := newJEX(file)
	if err != nil {
		return nil, err
	}

	return newJoplinFS(jex)
}

// Detect determines if the URL is likely to be a JEX file.
func Detect(dbURL *url.URL) notedb.DetectResult {
	if strings.HasSuffix(dbURL.Path, ".jex") {
		return notedb.DetectResultPositive
	}
	return notedb.DetectResultNegative
}
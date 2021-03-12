package jex

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/CGamesPlay/pilikino/pkg/markdown"
	"github.com/CGamesPlay/pilikino/pkg/pilikino"
	"github.com/Masterminds/sprig"
	"github.com/araddon/dateparse"
	"github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	gmparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func randomID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func tpl(name, content string) *template.Template {
	return template.Must(template.New(name).Funcs(sprig.TxtFuncMap()).Parse(content))
}

type objectInfo struct {
	ID          string
	ParentID    string
	Name        string
	CreatedTime time.Time
	UpdatedTime time.Time
}

type noteInfo struct {
	objectInfo
	Content string
}

type resourceInfo struct {
	objectInfo
	MimeType  string
	Extension string
	Size      int64
}

type tagLinkInfo struct {
	objectInfo
	TagID  string
	NoteID string
}

type export struct {
	index     *pilikino.Index
	indexDir  string
	targetDir string
	objects   map[string]*objectInfo
}

var tagSep = regexp.MustCompile("[^-a-z0-9_]+")

var notebookTpl = tpl("notebook", `{{.Name}}

id: {{.ID}}
created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
encryption_cipher_text: 
encryption_applied: 0
parent_id: {{.ParentID}}
is_shared: 0
type_: 2`)

var noteTpl = tpl("note", `{{.Name}}

{{.Content}}

id: {{.ID}}
parent_id: {{.ParentID}}
created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
is_conflict: 0
latitude: 0.00000000
longitude: 0.00000000
altitude: 0.0000
author: 
source_url: 
is_todo: 0
todo_due: 0
todo_completed: 0
source: pilikino
source_application: com.cgamesplay.pilikino
application_data: 
order: {{sub 0 .UpdatedTime.Unix}}
user_created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
encryption_cipher_text: 
encryption_applied: 0
markup_language: 1
is_shared: 0
type_: 1`)

var resourceTpl = tpl("resource", `{{.Name}}

id: {{.ID}}
mime: {{.MimeType}}
filename: 
created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
file_extension: {{.Extension}}
encryption_cipher_text: 
encryption_applied: 0
encryption_blob_encrypted: 0
size: {{.Size}}
is_shared: 0
type_: 4`)

var tagTpl = tpl("tag", `{{.Name}}

id: {{.ID}}
created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
encryption_cipher_text: 
encryption_applied: 0
is_shared: 0
parent_id: 
type_: 5`)

var tagLinkTpl = tpl("taglink", `id: {{.ID}}
note_id: {{.NoteID}}
tag_id: {{.TagID}}
created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_created_time: {{.CreatedTime.Format "2006-01-02T15:04:05Z07:00"}}
user_updated_time: {{.UpdatedTime.Format "2006-01-02T15:04:05Z07:00"}}
encryption_cipher_text: 
encryption_applied: 0
is_shared: 0
type_: 6`)

// Export parses all of the documents in the index and converts them into a RAW
// Joplin export directory located at targetDir.
func Export(index *pilikino.Index, targetDir string) error {
	if err := os.Mkdir(targetDir, os.ModePerm); err != nil {
		return err
	}
	indexDir, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	exp := export{
		index:     index,
		indexDir:  indexDir,
		targetDir: targetDir,
		objects:   map[string]*objectInfo{},
	}
	return exp.Run()
}

// Run does the full conversion.
func (exp *export) Run() error {
	err := filepath.Walk(exp.indexDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		baseName := filepath.Base(path)
		if path != exp.indexDir && strings.HasPrefix(baseName, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			err = exp.handleNotebook(path, info)
		} else if strings.HasSuffix(path, ".md") {
			err = exp.handleNote(path, info)
		} else {
			err = exp.handleResource(path, info)
		}
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (exp *export) getObject(path string) *objectInfo {
	oi, ok := exp.objects[path]
	if ok {
		return oi
	}
	oi = &objectInfo{
		ID: randomID(),
	}
	if path != exp.indexDir && path != "." {
		parent := exp.getObject(filepath.Dir(path))
		oi.ParentID = parent.ID
	}
	exp.objects[path] = oi
	return oi
}

func (exp *export) handleNotebook(path string, info os.FileInfo) error {
	oi := exp.getObject(path)
	oi.Name = filepath.Base(path)
	oi.CreatedTime = info.ModTime()
	oi.UpdatedTime = info.ModTime()

	f, err := os.Create(filepath.Join(exp.targetDir, oi.ID+".md"))
	defer f.Close()
	if err != nil {
		return err
	}
	if err := notebookTpl.Execute(f, oi); err != nil {
		return err
	}
	fmt.Printf("%v  %v\n", oi.ID, path)
	return err
}

func (exp *export) handleNote(path string, info os.FileInfo) error {
	oi := exp.getObject(path)
	oi.Name = filepath.Base(path)
	oi.CreatedTime = info.ModTime()
	oi.UpdatedTime = info.ModTime()

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	ni := &noteInfo{
		objectInfo: *oi,
	}

	err = exp.parseMarkdownNote(ni, path, bytes)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(exp.targetDir, oi.ID+".md"))
	defer f.Close()
	if err != nil {
		return err
	}
	if err := noteTpl.Execute(f, ni); err != nil {
		return err
	}
	fmt.Printf("%v  %v\n", oi.ID, path)
	return nil
}

func (exp *export) handleResource(path string, info os.FileInfo) error {
	if err := os.Mkdir(filepath.Join(exp.targetDir, "resources"), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}

	oi := exp.getObject(path)
	oi.Name = filepath.Base(path)
	oi.CreatedTime = info.ModTime()
	oi.UpdatedTime = info.ModTime()

	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if ext != "" {
		ext = ext[1:]
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	ri := &resourceInfo{
		objectInfo: *oi,
		Size:       info.Size(),
		Extension:  ext,
		MimeType:   mimeType,
	}

	src, err := os.Open(path)
	defer src.Close()
	if err != nil {
		return err
	}

	dst, err := os.Create(filepath.Join(exp.targetDir, "resources", oi.ID+"."+ext))
	defer dst.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(exp.targetDir, oi.ID+".md"))
	defer f.Close()
	if err != nil {
		return err
	}
	if err := resourceTpl.Execute(f, ri); err != nil {
		return err
	}
	fmt.Printf("%v  %v\n", oi.ID, path)
	return err
}

func (exp *export) handleTag(name string) (*objectInfo, error) {
	oi := exp.getObject("#" + name)
	if oi.Name != "" {
		return oi, nil
	}
	oi.Name = name
	oi.CreatedTime = time.Now()
	oi.UpdatedTime = time.Now()

	f, err := os.Create(filepath.Join(exp.targetDir, oi.ID+".md"))
	defer f.Close()
	if err != nil {
		return nil, err
	}
	if err := tagTpl.Execute(f, oi); err != nil {
		return nil, err
	}
	fmt.Printf("%v  %v\n", oi.ID, "#"+name)
	return oi, nil
}

func (exp *export) handleTagLink(ni *noteInfo, ti *objectInfo) error {
	oi := exp.getObject(ni.ID + ":" + ti.ID)
	oi.CreatedTime = time.Now()
	oi.UpdatedTime = time.Now()

	tli := &tagLinkInfo{
		objectInfo: *oi,
		TagID:      ti.ID,
		NoteID:     ni.ID,
	}

	f, err := os.Create(filepath.Join(exp.targetDir, oi.ID+".md"))
	defer f.Close()
	if err != nil {
		return err
	}
	if err := tagLinkTpl.Execute(f, tli); err != nil {
		return err
	}
	fmt.Printf("%v  %v\n", oi.ID, ni.Name+": #"+ti.Name)
	return nil
}

func (exp *export) parseMarkdownNote(ni *noteInfo, path string, file []byte) error {
	parser := goldmark.New(goldmark.WithExtensions(meta.Meta, extension.GFM, mathjax.MathJax))
	context := gmparser.NewContext()
	reader := text.NewReader(file)
	doc := parser.Parser().Parse(reader, gmparser.WithContext(context))
	frontmatter := meta.Get(context)
	if titleVal, ok := frontmatter["title"]; ok {
		if titleStr, ok := titleVal.(string); ok {
			ni.Name = titleStr
		}
	}
	if ctimeVal, ok := frontmatter["date"]; ok {
		if ctimeStr, ok := ctimeVal.(string); ok {
			ctime, err := dateparse.ParseLocal(ctimeStr)
			if err == nil {
				ni.CreatedTime = ctime
			}
		}
	}
	if tagField, ok := frontmatter["tags"]; ok {
		var tagList []string
		if tagStr, ok := tagField.(string); ok {
			tagList = tagSep.Split(tagStr, -1)
		} else if tagListVar, ok := tagField.([]interface{}); ok {
			tagList = make([]string, len(tagListVar))
			for i, v := range tagListVar {
				tagList[i], _ = v.(string)
			}
		}
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				ti, err := exp.handleTag(tag)
				if err != nil {
					return err
				}
				err = exp.handleTagLink(ni, ti)
				if err != nil {
					return err
				}
			}
		}
	}

	converter := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindLink:
			link := n.(*ast.Link)
			dest := string(link.Destination)
			resolved, err := exp.index.ResolveLink(path, dest)
			if err != nil {
				return ast.WalkContinue, nil
			}
			resolved = filepath.Join(exp.indexDir, resolved)
			oi := exp.getObject(resolved)
			link.Destination = []byte(":/" + oi.ID)
		case ast.KindImage:
			image := n.(*ast.Image)
			dest := string(image.Destination)
			resolved, err := exp.index.ResolveLink(path, dest)
			if err != nil {
				return ast.WalkContinue, nil
			}
			resolved = filepath.Join(exp.indexDir, resolved)
			oi := exp.getObject(resolved)
			image.Destination = []byte(":/" + oi.ID)
		}
		return ast.WalkContinue, nil
	}

	if err := ast.Walk(doc, converter); err != nil {
		return err
	}

	b := &strings.Builder{}
	err := markdown.NewRenderer().Render(b, file, doc)
	if err != nil {
		return err
	}
	ni.Content = b.String()
	return nil
}

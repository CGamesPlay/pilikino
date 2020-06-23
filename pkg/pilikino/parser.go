package pilikino

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/blevesearch/bleve"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type noteData struct {
	Filename    string    `json:"filename"`
	Date        time.Time `json:"date"`
	ModTime     time.Time `json:"modified"`
	Title       string    `json:"title"`
	Tags        []string  `json:"tags"`
	Links       []string  `json:"links"`
	Content     string    `json:"content"`
	ParseErrors []string  `json:"errors"`
}

func (note *noteData) AddLink(dest string) {
	note.Links = append(note.Links, dest)
}

func (note *noteData) AddParseError(err string) {
	note.ParseErrors = append(note.ParseErrors, err)
}

var tagSep = regexp.MustCompile("[^-a-z0-9_]+")

func resolveLink(index *Index, from, to string) (string, error) {
	toURI, err := url.Parse(to)
	if err != nil {
		return "", err
	}
	if toURI.Host != "" {
		return "", nil
	}
	return index.ResolveLink(from, to)
}

func extractLinks(index *Index, note *noteData, file []byte) func(n ast.Node, entering bool) (ast.WalkStatus, error) {
	return func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindLink {
			link, ok := n.(*ast.Link)
			if !ok {
				return ast.WalkStop, errors.New("node KindLink is not ast.Link")
			}
			dest := string(link.Destination)
			resolved, err := resolveLink(index, note.Filename, dest)
			if err != nil {
				note.AddParseError(fmt.Sprintf("%v: %#v", err, dest))
			} else {
				note.AddLink(resolved)
			}
		}
		return ast.WalkContinue, nil
	}
}

func parseMarkdownNote(index *Index, note *noteData, file []byte) error {
	markdown := goldmark.New(goldmark.WithExtensions(meta.Meta))
	context := parser.NewContext()
	reader := text.NewReader(file)
	doc := markdown.Parser().Parse(reader, parser.WithContext(context))
	frontmatter := meta.Get(context)
	if titleVal, ok := frontmatter["title"]; ok {
		if titleStr, ok := titleVal.(string); ok {
			note.Title = titleStr
		} else {
			note.AddParseError("title is not string")
		}
	}
	if ctimeVal, ok := frontmatter["date"]; ok {
		if ctimeStr, ok := ctimeVal.(string); ok {
			ctime, err := dateparse.ParseLocal(ctimeStr)
			if err != nil {
				note.AddParseError(fmt.Sprintf("ctime unrecognized: %v", err))
			} else {
				note.Date = ctime
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
		note.Tags = []string{}
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				note.Tags = append(note.Tags, tag)
			}
		}
	}
	if err := ast.Walk(doc, extractLinks(index, note, file)); err != nil {
		note.AddParseError(err.Error())
	}
	return nil
}

func indexNote(index *Index, batch *bleve.Batch, path string, info os.FileInfo) error {
	if !strings.HasSuffix(path, ".md") {
		return nil
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	note := noteData{
		Filename: path,
		ModTime:  info.ModTime(),
		Title:    path,
		Content:  string(content),
	}
	err = parseMarkdownNote(index, &note, content)
	if indexErr := batch.Index(path, note); indexErr != nil {
		note.ParseErrors = append(note.ParseErrors, err.Error())
	}
	return err
}

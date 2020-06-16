package pilikino

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/blevesearch/bleve"
	"gopkg.in/yaml.v3"
)

type noteData struct {
	Filename    string    `json:"filename"`
	CreateTime  time.Time `json:"created"`
	ModTime     time.Time `json:"modified"`
	Title       string    `json:"title"`
	Tags        []string  `json:"tags"`
	Content     string    `json:"content"`
	ParseErrors []string  `josn:"errors"`
}

var tagSep = regexp.MustCompile("[^-a-z0-9_]+")

func parseHeader(note *noteData, header map[string]interface{}) error {
	if title, ok := header["title"]; ok {
		note.Title = title.(string)
	}
	if ctimeVal, ok := header["date"]; ok {
		if ctimeStr, ok := ctimeVal.(string); ok {
			ctime, err := dateparse.ParseLocal(ctimeStr)
			if err != nil {
				note.ParseErrors = append(note.ParseErrors, err.Error())
			} else {
				note.CreateTime = ctime
			}
		}

	}
	if tagField, ok := header["tags"]; ok {
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
	return nil
}

func parseNoteContent(note *noteData) error {
	b := []byte(note.Content)
	if bytes.HasPrefix(b, []byte("---\n")) {
		endHeader := bytes.Index(b, []byte("\n---\n"))
		if endHeader == -1 {
			return nil
		}
		header := make(map[string]interface{})
		if err := yaml.Unmarshal(b[0:endHeader], &header); err != nil {
			return err
		}
		if err := parseHeader(note, header); err != nil {
			return err
		}
	}
	return nil
}

func convertNewlines(b []byte) []byte {
	b = bytes.Replace(b, []byte{'\r', '\n'}, []byte{'\n'}, -1)
	b = bytes.Replace(b, []byte{'\r'}, []byte{'\n'}, -1)
	return b
}

func indexNote(batch *bleve.Batch, path string, info os.FileInfo) error {
	if !strings.HasSuffix(path, ".md") {
		return nil
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	content = convertNewlines(content)
	note := noteData{
		Filename:    path,
		ModTime:     info.ModTime(),
		Title:       path,
		Content:     string(content),
		ParseErrors: []string{},
	}
	err = parseNoteContent(&note)
	if indexErr := batch.Index(path, note); indexErr != nil {
		note.ParseErrors = append(note.ParseErrors, err.Error())
	}
	return err
}

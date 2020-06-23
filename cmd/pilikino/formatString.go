package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	bleveSearch "github.com/blevesearch/bleve/search"
	"github.com/spf13/cobra"
)

var resultTemplateStr = `{{.id}}`
var resultTemplate *template.Template

func init() {
	rootCmd.AddCommand(formatStringCmd)
}

var formatStringCmd = &cobra.Command{
	Use:   "format-strings",
	Short: "Help about format strings",
	Long: `Pilikino uses go template strings for formatting results. A full reference about go format strings is available here:
https://golang.org/pkg/text/template/

Additionally, a function ` + "`json`" + ` is provided to JSON-encode values.

Examples:
- The default template: ` + "`" + resultTemplateStr + "`" + `
- To print the filename and title: ` + "`{{.id}}: {{.hit.title}}`" + `
- To print the first match fragment as JSON: ` + "`{{json (index .hit.fragments 0)}}`" + `
- To print everything as json: ` + "`{{json .}}`",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Long)
	},
}

func setupResultTemplate() error {
	if resultTemplate != (*template.Template)(nil) {
		return nil
	}
	funcMap := make(map[string]interface{})
	funcMap["json"] = func(v interface{}) (string, error) {
		a, err := json.Marshal(v)
		return string(a), err
	}
	t := template.New("")
	t.Funcs(funcMap)
	t, err := t.Parse(resultTemplateStr)
	if err != nil {
		return err
	}
	resultTemplate = t
	return nil
}

func formatResult(hit *bleveSearch.DocumentMatch) (string, error) {
	if err := setupResultTemplate(); err != nil {
		return "", err
	}
	var result strings.Builder
	data := make(map[string]interface{})
	data["id"] = hit.ID
	data["score"] = hit.Score
	data["hit"] = hit.Fields
	data["fragments"], _ = hit.Fragments["content"]
	if err := resultTemplate.Execute(&result, data); err != nil {
		return "", err
	}
	return result.String(), nil
}

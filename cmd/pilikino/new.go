package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

var newFileTags []string

func init() {
	newCmd.Flags().StringSliceVar(&newFileTags, "tags", []string{}, "list of tags to add to the frontmatter")
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new TITLE",
	Short: "Create a new note",
	Long:  `Creates a new file with the provided title. The configuration file specifies how the file is named and its initial content..`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := getIndex()
		checkError(err)

		info := &NewFileInfo{
			Title: strings.Join(args, " "),
			Date:  time.Now(),
			Tags:  newFileTags,
		}
		result, err := newCreateResult(info)
		checkError(err)
		filename, err := result.Create()
		checkError(err)
		fmt.Printf("%v\n", filename)
	},
}

type createResult struct {
	filename string
	content  string
}

func newCreateResult(info *NewFileInfo) (*createResult, error) {
	if info.Date == (time.Time{}) {
		info.Date = time.Now()
	}
	filename, err := DefaultConfig.GetFilename(info)
	if err != nil {
		return nil, err
	}
	content, err := DefaultConfig.GetTemplate(info)
	if err != nil {
		return nil, err
	}
	result := &createResult{filename, content}
	return result, nil
}

func (hit *createResult) Label() string {
	return fmt.Sprintf("Create %#v", hit.filename)
}

func (hit *createResult) Preview(preview *tview.TextView) {
	preview.SetScrollable(true).
		SetText(hit.content).
		SetWordWrap(true).
		SetDynamicColors(false).
		ScrollTo(0, 0)
}

func (hit *createResult) Create() (string, error) {
	err := ioutil.WriteFile(hit.filename, []byte(hit.content), 0644)
	return hit.filename, err
}

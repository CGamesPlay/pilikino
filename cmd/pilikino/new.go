package main

import (
	"fmt"
	"os"
	"strings"
	"time"

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
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(ExitStatusError)
		}

		info := &NewFileInfo{
			Title: strings.Join(args, " "),
			Date:  time.Now(),
			Tags:  newFileTags,
		}
		filename, err := DefaultConfig.GetFilename(info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(ExitStatusError)
		}
		content, err := DefaultConfig.GetTemplate(info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(ExitStatusError)
		}
		fmt.Printf("Create a new file %#v!\n", filename)
		fmt.Printf("%v\n", content)
	},
}

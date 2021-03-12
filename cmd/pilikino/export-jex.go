package main

import (
	"path/filepath"

	"github.com/CGamesPlay/pilikino/pkg/jex"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exportJexCmd)
}

var exportJexCmd = &cobra.Command{
	Use:   "export-jex DIRECTORY",
	Short: "Export all notes to a RAW Joplin export directory.",
	Long:  `Converts all notes and attachments into a format suitable for exporting into Joplin.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetDir, err := filepath.Abs(args[0])
		checkError(err)
		index, err := getIndex()
		checkError(err)

		err = jex.Export(index, targetDir)
		checkError(err)
	},
}

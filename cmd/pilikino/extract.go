package main

import (
	"io"
	"os"

	"github.com/CGamesPlay/pilikino/lib/notedb"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "extract DATABASE PATH",
		Short: "Extract a single note from the database",
		Long:  `Print out the Markdown source of a note (or binary data of an attachment).`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			dbURL, err := notedb.ResolveURL(args[0])
			if err != nil {
				exitError(1, "Cannot determine database type: %s\n", err)
			}
			db, err := notedb.OpenDatabase(dbURL)
			if err != nil {
				exitError(1, "Cannot open database: %s\n", err)
			}

			file, err := db.Open(args[1])
			if err != nil {
				exitError(1, "%s\n", err)
			}

			_, err = io.Copy(os.Stdout, file)
			if err != nil {
				exitError(1, "%s\n", err)
			}
		},
	}
	rootCmd.AddCommand(cmd)
}

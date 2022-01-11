package main

import (
	"fmt"
	"io/fs"

	"github.com/CGamesPlay/pilikino/lib/notedb"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "ls DATABASE",
		Short: "List the contents of a database",
		Long:  `Produce a full listing of all of the notes and attachments in the database.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dbURL, err := notedb.ResolveURL(args[0])
			if err != nil {
				exitError(1, "Cannot determine database type: %s\n", err)
			}
			db, err := notedb.OpenDatabase(dbURL)
			if err != nil {
				exitError(1, "Cannot open database: %s\n", err)
			}

			err = fs.WalkDir(db, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				fmt.Printf("%s\n", path)
				return nil
			})
			if err != nil {
				exitError(1, "Cannot read database: %s\n", err)
			}
		},
	}
	rootCmd.AddCommand(cmd)
}

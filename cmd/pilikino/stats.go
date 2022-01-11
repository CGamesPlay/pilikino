package main

import (
	"fmt"
	"io/fs"

	"github.com/CGamesPlay/pilikino/lib/notedb"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "stats DATABASE",
		Short: "Print some statistics about a note database",
		Long:  `Fully loads all notes in the database, parses them, and produces some statistics.`,
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
				stat, err := d.Info()
				if err != nil {
					return err
				}
				noteStat, ok := stat.(notedb.FileInfo)
				if ok {
					if noteStat.IsNote() {
						fmt.Printf("%s\n", path)
					} else {
						fmt.Printf("%s\n", path)
					}
				}
				return nil
			})
			if err != nil {
				exitError(1, "Cannot read database: %s\n", err)
			}
		},
	}
	rootCmd.AddCommand(cmd)
}

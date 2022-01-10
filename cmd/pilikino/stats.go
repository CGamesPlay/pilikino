package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/CGamesPlay/pilikino/lib/notedb"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Print some statistics about a note database",
		Long:  `Fully loads all notes in the database, parses them, and produces some statistics.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dbURL, err := notedb.ResolveURL(args[0])
			if err != nil {
				fmt.Printf("Cannot determine database type: %s\n", err)
				os.Exit(1)
			}
			db, err := notedb.OpenDatabase(dbURL)
			if err != nil {
				fmt.Printf("Cannot open database: %s\n", err)
				os.Exit(1)
			}

			err = fs.WalkDir(db, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				fmt.Printf("%s\n", path)
				return nil
			})
			if err != nil {
				fmt.Printf("Cannot read database: %s\n", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(cmd)
}

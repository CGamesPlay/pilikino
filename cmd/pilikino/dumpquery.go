package main

import (
	"fmt"

	"github.com/blevesearch/bleve/search/query"
	"github.com/spf13/cobra"
)

var dumpQueryInteractive bool

func init() {
	searchCmd.Flags().BoolVar(&dumpQueryInteractive, "interactive", false, "treat query as interactive")
	rootCmd.AddCommand(dumpQueryCmd)
}

var dumpQueryCmd = &cobra.Command{
	Use:   "dumpquery QUERY",
	Short: "Shows the internal representation of the query",
	Run: func(cmd *cobra.Command, args []string) {
		index, err := getIndex()
		checkError(err)
		q, err := parseQuery(args[0], dumpQueryInteractive)
		checkError(err)
		res, err := query.DumpQuery(index.Bleve.Mapping(), q)
		checkError(err)
		fmt.Printf("%v\n", res)
	},
}

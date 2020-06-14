package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CGamesPlay/pilikino/pkg/pilikino"
	"github.com/blevesearch/bleve"
	"github.com/spf13/cobra"
)

const numResults = 100

var ErrNoResults = errors.New("no results")

func init() {
	rootCmd.AddCommand(filterCmd)
}

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Run a single query and print results",
	Long:  `Runs a query and outputs the results in the format specified. Designed to be used as input to other programs.`,
	Run: func(cmd *cobra.Command, args []string) {
		index, err := buildIndex()
		if err == nil {
			err = runFilter(index, strings.Join(args, " "))
		}

		// These exit codes are similar to grep's.
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		} else if err == ErrNoResults {
			// No matches
			os.Exit(1)
		}
	},
}

func runFilter(index *pilikino.Index, queryString string) error {
	query, err := parseQuery(queryString)
	if err != nil {
		return err
	}
	search := bleve.NewSearchRequestOptions(query, numResults, 0, false)
	search.Highlight = bleve.NewHighlight()
	res, err := index.Bleve.Search(search)
	if err != nil {
		return err
	} else if res.Total == 0 {
		return ErrNoResults
	}
	for _, hit := range res.Hits {
		fmt.Printf("%v\n", hit.ID)
	}
	return nil
}

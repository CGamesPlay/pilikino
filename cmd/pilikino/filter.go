package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/CGamesPlay/pilikino/pkg/pilikino"
	"github.com/spf13/cobra"
)

const numResults = 100

var errNoResults = errors.New("no results")

func init() {
	filterCmd.Flags().StringVarP(&resultTemplateStr, "format", "f", resultTemplateStr, "format string to use for results")
	rootCmd.AddCommand(filterCmd)
}

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Run a single query and print results",
	Long:  `Runs a query and outputs the results in the format specified. Designed to be used as input to other programs.`,
	Run: func(cmd *cobra.Command, args []string) {
		index, err := getIndex()
		if err == nil {
			err = index.Reindex(nil)
		}
		if err == nil {
			err = runFilter(index, strings.Join(args, " "))
		}

		// These exit codes are similar to grep's.
		if err == errNoResults {
			// No matches
			os.Exit(1)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
	},
}

func runFilter(index *pilikino.Index, queryString string) error {
	query, err := parseQuery(queryString, false)
	if err != nil {
		return err
	}
	res, err := performSearch(index, query, numResults)
	if err != nil {
		return err
	} else if res.Total == 0 {
		return errNoResults
	}
	for _, hit := range res.Hits {
		res, err := formatResult(hit)
		if err != nil {
			return err
		}
		fmt.Println(res)
	}
	return nil
}

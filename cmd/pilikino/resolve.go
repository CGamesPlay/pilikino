package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var resolveFrom string

func init() {
	resolveCmd.Flags().StringVarP(&resolveFrom, "from", "f", "", "filename to resolve relative from")
	rootCmd.AddCommand(resolveCmd)
}

var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve a relative link",
	Long:  `Prints the resolved filename for the given URL.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var result string
		index, err := getIndex()
		if err == nil {
			result, err = index.ResolveLink(resolveFrom, args[0])
		}

		checkError(err)
		if result == "" {
			fmt.Fprintf(os.Stderr, "%v\n", errNoResults)
			os.Exit(ExitStatusNoResults)
		} else {
			fmt.Println(result)
		}
	},
}

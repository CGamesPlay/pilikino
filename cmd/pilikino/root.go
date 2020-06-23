package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var directory string

var rootCmd = &cobra.Command{
	Use:  "pilikino",
	Long: "Pilikino helps you find where you wrote that idea down.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If the root command is run directly, invoke the search command
		// instead.
		// TODO - this only works when there are no arguments passed, since
		// they would be rejected before reaching this point.
		args = []string{searchCmd.Name()}
		cmd.SetArgs(args)
		return cmd.Execute()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&directory, "directory", "C", "", "change into directory before starting")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitStatusError)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newConfigCmd)
}

var newConfigCmd = &cobra.Command{
	Use:   "newconfig",
	Short: "Create a new, default config file",
	Long:  `Creates a pilikino configuration file if none exists.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := getIndex(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(ExitStatusError)
		}
		if err := DefaultConfig.Save(false); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(ExitStatusError)
		}
	},
}

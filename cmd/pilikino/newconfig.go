package main

import (
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
		_, err := getIndex()
		checkError(err)
		checkError(DefaultConfig.Save(false))
	},
}

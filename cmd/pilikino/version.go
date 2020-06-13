package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current version of the program.
const Version = "1.0.0"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("Prints the version (%v)", Version),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Pilikino v%v\n", Version)
	},
}

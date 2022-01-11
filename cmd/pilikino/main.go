package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/CGamesPlay/pilikino/lib/formats/file"
	_ "github.com/CGamesPlay/pilikino/lib/formats/jex"
)

var rootCmd = &cobra.Command{
	Use:   "pilikino",
	Short: "Pilikino is a Swiss Army knife for note taking apps",
	Long:  `Provides a set of tools to import, export, view, and modify collections of notes.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func logError(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, format, a...)
}

func exitError(exitCode int, format string, a ...interface{}) {
	logError(format, a...)
	os.Exit(exitCode)
}

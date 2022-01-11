package main

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/CGamesPlay/pilikino/lib/notedb"
	fs "github.com/relab/wrfs"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

func init() {
	cmd := &cobra.Command{
		Use:   "convert SOURCE DEST",
		Short: "Convert an entire database from one format to another.",
		Long:  `Convert an entire database from one format to another.`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			srcURL, err := notedb.ResolveURL(args[0])
			if err != nil {
				exitError(1, "Cannot determine database type: %s\n", err)
			}
			src, err := notedb.OpenDatabase(srcURL)
			if err != nil {
				exitError(1, "Cannot open source database: %s\n", err)
			}

			dstURL, err := notedb.ResolveURL(args[1])
			if err != nil {
				exitError(1, "Cannot determine database type: %s\n", err)
			}
			dst, err := notedb.OpenDatabase(dstURL)
			if err != nil {
				exitError(1, "Cannot open destination database: %s\n", err)
			}

			var allErrs error

			err = fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}

				var fileErrs error
				defer func() {
					if fileErrs != nil {
						errs := multierr.Errors(fileErrs)
						for _, err := range errs {
							allErrs = multierr.Append(allErrs, &fs.PathError{Op: "convert", Path: path, Err: err})
						}
					}
				}()

				srcFile, err := src.Open(path)
				if err != nil {
					return err
				}
				defer srcFile.Close()

				base, _ := filepath.Split(path)
				if base != "" {
					base = base[:len(base)-1]
					if err := fs.MkdirAll(dst, base, 0777); err != nil {
						return err
					}
				}
				dstFile, err := fs.Create(dst, path)
				if err != nil {
					return err
				}

				if note, ok := srcFile.(notedb.Note); ok && note.IsNote() {
					ast, err := note.ParseAST()
					if err != nil {
						fileErrs = multierr.Append(fileErrs, err)
					}
					if dstNote, ok := dstFile.(notedb.Note); ok {
						if err := notedb.WriteAST(dstNote, ast, note.Data()); err != nil {
							fileErrs = multierr.Append(fileErrs, err)
						}
					} else {
						// This is a fatal error for this file, cover up
						// everything else.
						fileErrs = fmt.Errorf("destination file is not a note")
					}
				} else {
					_, fileErrs = io.Copy(dstFile, srcFile)
				}

				dstFile.Close()

				stat, err := d.Info()
				if err != nil {
					fileErrs = multierr.Append(fileErrs, err)
				} else if err = fs.Chtimes(dst, path, time.Now(), stat.ModTime()); err != nil {
					fileErrs = multierr.Append(fileErrs, err)
				}

				return nil
			})
			if err != nil {
				exitError(1, "Error reading database: %s\n", err)
			}

			errs := multierr.Errors(allErrs)
			logError("Finished with %d errors\n", len(errs))
			for _, err := range errs {
				logError("%s\n", err)
			}
		},
	}
	rootCmd.AddCommand(cmd)
}

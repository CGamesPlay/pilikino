# Pilikino - personal wiki tools

Pilikino is a full-text search tool for navigating a directory of interlinked Markdown files.

What are the tasks that this tool will perform?

- Full-text search of notes (CLI, vim, Alfred)
- Find backlinks / "related pages" (backlinks, forward links, document similarity?)
- Planned: Find orphans
- Planned: Format migrations (e.g. relative links <-> wikilinks)
- Resolve ambiguous link (e.g. numeric file prefix instead of full link)

## Installation

Pilikino is under active development. You can `go get` it, but binary releases are not yet available. Roadmap and general development notes are available in [TODO.md](TODO.md).

## Usage

### Interactive search

The main mode of Pilikino is the interactive search. It operates similarly to [fzf](https://github.com/junegunn/fzf), where you select a file interactively, and Pilikino prints the filename that you selected.

```bash
# Select a file in the current directory, and open it in vim.
vim `pilikino`
```

Pilikino always operates on files in the current directory, but you can instruct the program to change into a directory before scanning:

```bash
pilikino search -C ~/notes
```

### Vim plugin

The `vim` directory of this repository contains a vim plugin that adds a command, `:Pilikino`, which will open Pilikino in `g:pilikino_directory`, and edit any file that you select.

This plugin is currently under active development, so it's not yet compatible with plugin managers.

### Alfred workflow

The included [Alfred](https://www.alfredapp.com) workflow, coupled with [Typora](https://typora.io) or another suitable Markdown editor, work well together to create a personal wiki. Once you have installed the workflow and set the `DIR` variable (you may also need to set `PATH`), you will be able to:

Open a note from anywhere:

- Press F3 in any app to open the search interface, prefilled with the current Mac OS selection.
- Alfred results will include the relevant part of the note.
- Tap Shift to show the document in Quicklook (install [QLMarkdown](https://github.com/toland/qlmarkdown)).
- Press Enter to open the result in the default Markdown editor (set this to be Typora).

Link to another note:

- Press F3 in the editor to open the search interface, prefilled with the currently selected text.
- Press Command-C to copy a Markdown link to the result.
- Press Command-V to replace the currently selected text with the Markdown link.

Navigate between notes (these are Typora tricks; not the Alfred workflow):

- Command-Click on a link to another note to open it in Typora.
- Press Command-T to open a new tab before opening a note to have the note opened in that tab instead of a new window.

## Reference

### Query syntax

Currently supported query syntax:

- basic english queries `recent changes` will search for "recent" and "changes", with stemming (so it might match "recently changed").
- phrase queries `"recent changes"` will perform an exact phrase match for "recent changes", with stemming (so it will match "recently changed" but not "recent organizational changes").
- literal queries \`docker run\` will perform an exact search, without stemming. (:bug: exact terms aren't indexed properly so this rarely works)
- field queries `title:bleve` will search only in not titles. Available fields include "content", "title", "tags", and "links"
- backlinks queries `links:exact-filename.md` will list all documents that link to the given one
- hashtags `#cooking` will list documents that have that tag in the YAML front-matter of the document (shortcut for `tags:cooking`)

### Other

Pilikino will exit with 0 when the requested operation is successful. An exit status of 1 indicates that the program operated normally, but no results for the query were found. An exit status of 130 indicates that the user aborted an interactive search ([why?](https://neovim.io/doc/user/job_control.html#on_exit)). For all other errors, the program exits with status 2.

## Known issues

- Flicker? Caused by [rivo/tview#314](https://github.com/rivo/tview/issues/314).


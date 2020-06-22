# Pilikino - personal wiki tools

This is a quick demo of bleve search on a directory of Markdown files.

What are the tasks that this tool will perform?

- Full-text search of notes (CLI, vim, Alfred)
- Find backlinks / "related pages" (backlinks, forward links, document similarity?)
- Find orphans
- Format migrations (e.g. relative links <-> wikilinks)
- Resolve ambiguous link (e.g. numeric file prefix instead of full link)

### Current task

Nothing!

### Next steps

- More vim polish
- Improve match highlighting preview window
- Tune boost values

### Vim plugin TODO list

- Write documentation and vroom tests
- Split into a basic search (interactive and non interactive) interface and rest
- Create a non interactive mode for use with backlinks search

### Querying TODO list

- `+` should remove the default matchall
- improve being-typed word matching and remove the default match-all
- filename search
- automatic quote closing
- date searching for created / modified
- convenience searches for `has:errors`, `is:orphan`, etc
- probably just stop using yacc

### Improving highlighting

In order to identify locations of links in documents, the pipeline will look like this:

- Character filter to remove everything except markdown link source at same position in document
- Whitespace tokenizer
- Token filter to replace the extracted links with resolved link hrefs.

The result is that `links` now becomes a text field with this analyzer instead of an array. Highlighting uses offsets from `links` but projected onto the `content` field. We can easily extract line numbers of links. This same process will apply to inline tags as well.

To highlight this, we have to copy the `highlighter_simple.go` code, but we can mostly reuse it. We need to change the input value to the fragmenter to receive the original content value instead of the filtered value.

## Reference

### Query syntax

What should be queryable? tags, text, filenames, titles, dates, arbitrary front matter, backlinks, forward links. Some niceties like "orphan" (no backlinks), configuring sort order.

- basic english queries `terrain generation`
- phrase queries `"terrain generation"`
- literal queries \`docker run\`
- negation `-docker` or `not docker`
- parenthesis `(docker kubernetes)`
- booleans `docker and kubernetes` `docker or kubernetes`
- field searches `filename:blah`
- presence: `has:field`
- date queries: `created > "1 week ago"`

In interactive mode:

- closing parentheses, quotes, etc are automatically balanced at the end of the query
- the final term can be matched by prefix only

### Alfred workflow

The included [Alfred](https://www.alfredapp.com) workflow, coupled with [Typora](https://typora.io) or another suitable Markdown editor, work well together to create a personal wiki. Once you have installed the workflow and set the `DIR` variable (you may also need to set `PATH`), you will be able to:

Open a note from anywhere:

- Press F3 anywhere to open the search interface, prefilled with the current Mac OS selection.
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

### Other

Exit status:

- 0 success
- 1 no match
- 2 error
- 130 interrupted

## Known issues

- Flicker? Caused by [rivo/tview#314](https://github.com/rivo/tview/issues/314).


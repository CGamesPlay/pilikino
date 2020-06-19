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
- Alfred workflow
- Improve match highlighting preview window

### Vim plugin TODO list

- Write documentation and vroom tests
- Write `pilikino#writing` optional plugin and move link insertion, following there.

### Querying TODO list

- `+` should remove the default matchall
- improve being-typed word matching and remove the default match-all
- filename search
- automatic quote closing
- date searching for created / modified
- convenience searches for `has:errors`, `is:orphan`, etc
- probably just stop using yacc

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

### Other

Exit status:

- 0 success
- 1 no match
- 2 error
- 130 interrupted

## Known issues

- Flicker? Caused by [rivo/tview#314](https://github.com/rivo/tview/issues/314).


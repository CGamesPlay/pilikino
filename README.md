# Pilikino - personal wiki tools

This is a quick demo of bleve search on a directory of Markdown files.

### Current task

Command line interface.

What are the tasks that this tool will perform?

- Full-text search of notes (CLI, vim, Alfred)
- Find backlinks / "related pages" (backlinks, forward links, document similarity?)
- Find orphans
- Format migrations (e.g. relative links <-> wikilinks)
- Resolve ambiguous link (e.g. numeric file prefix instead of full link)

### Next steps

- Write a better command line interface
- Vim plugin polish (vertical split, etc). Goal: replace existing vimwiki config
- Improve search functionality (queries, ranking, recency)
- Indexing in background
- Investigate Maktaba, Glaive, Vroom for vim plugin

## Reference

Exit status:

- 0 success
- 1 no match
- 2 error
- 130 interrupted

## Known issues

- Flicker? Caused by [rivo/tview#314](https://github.com/rivo/tview/issues/314).


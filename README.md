# Pilikino - personal wiki tools

This is a quick demo of bleve search on a directory of Markdown files.

### Current task

Improve crawling/parsing process

- Async crawling
- indexing frontmatter
- indexing link graph

### Next steps

- Improve search functionality (queries, ranking, recency)
- More vim polish
- Alfred workflow
- Indexing in background

What are the tasks that this tool will perform?

- Full-text search of notes (CLI, vim, Alfred)
- Find backlinks / "related pages" (backlinks, forward links, document similarity?)
- Find orphans
- Format migrations (e.g. relative links <-> wikilinks)
- Resolve ambiguous link (e.g. numeric file prefix instead of full link)

### Vim plugin TODO list

- Write documentation and vroom tests
- Write `pilikino#writing` optional plugin and move link insertion, following there.

## Reference

Exit status:

- 0 success
- 1 no match
- 2 error
- 130 interrupted

## Known issues

- Flicker? Caused by [rivo/tview#314](https://github.com/rivo/tview/issues/314).


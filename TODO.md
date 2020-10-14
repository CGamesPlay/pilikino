# Pilikino TODO

This document tracks my own notes and thoughts about the project. Anything in this document may not reflect the actual state of the repository, or even my definite plans.

## Active development

(This is my personal TODO list for this project)

### Current task

Nothing!

### Next steps

- Allow following links in vim
- Orphan search - identify notes not linked from any others
- Format migrations
- Tool to dump parsed representation of document for debugging
- Graph visualization
- More vim polish
- Improve match highlighting preview window

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
- Synonyms dictionary (e.g. "prod" = "production")

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

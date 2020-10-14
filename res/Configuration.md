# Pilikino Configuration

Pilikino will read a file called `.pilikino.toml` in the current directory (or the directory specified with `-C`). This file can be used to alter the behavior of pilikino.

To create a config file in the appropriate location, run `pilikino newconfig`. This command will fail if an existing configuration file already exists. Here is an example of the default configuration file:

```toml
# defines the filename of new notes
filename = "{{.Date | date \"20060102-1504\"}}-{{.Title | kebabcase}}.md"

# defines the content of new notes
template = """
---
title: {{.Title}}
date: {{.Date | date "2006-01-02 15:04"}}
tags: {{.Tags | join " "}}
---

"""
```

## New file templates

When running `pilikino new`, pilikino creates a file based off of two templates.

- `filename` is used to decide where to place the file and how it will be named.

- `template` is used to determine the initial contents of the created file.

These templates are both [standard go templates](https://golang.org/pkg/text/template/) with [functions from sprig](https://masterminds.github.io/sprig/) installed. The variables usable by the template are:

- `.Title` - a string containing all of the arguments passed to `pilikino new` joined by spaces.
- `.Date` - the current date.
- `.Tags` - a list of strings corresponding to the `--tags` argument to `pilikino new`.
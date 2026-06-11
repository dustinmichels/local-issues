# Issues

Local issue tracking, using the [li](https://github.com/dustinmichels/local-issues) CLI.
It is designed to be read and edited by both humans and coding agents.

## Layout

- `NNN.toml` - open or in-progress issues
- `done/NNN.toml` - issues that have been finished

## Issue file format

```toml
[issue]
id = 1
title = "Example issue"
status = "open" # [open, in-progress, done]
assigned_to = ""
created_at = 2024-01-01T00:00:00Z

[details]
description = """
What needs to happen, and why.
"""

[subtasks]
"subtask 1" = false
"subtask 2" = true

[resolution]
notes = ""
```

When a task is completed, add `completed_at` and record `resolution.notes`.

## CLI commands

Run from anywhere inside the project; `li` walks up from the
current directory to find `.issues/`, the same way git finds `.git/`.

```sh
# create a new open issue
li create -t "Title" -d "Description" \
  --subtask "step one" --subtask "step two"

# list all issues
li list

# list issues with a given status (open, in-progress, done)
li list --status open

# show the title, description, path, and subtasks for one issue
li get <id>

# mark an issue done, record resolution notes, and move it to done/
li finish <id> --notes "Resolution notes"

# move any issues already marked done into done/
li cleanup

# serve a local web UI for browsing issues
li serve
```

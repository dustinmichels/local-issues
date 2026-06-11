# Issues

This directory is managed by [local-issues](https://github.com/dustinmichels/local-issues),
a CLI for tracking issues as local TOML files. It is designed to be read and
edited by both humans and coding agents.

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

`status` and `assigned_to` can be edited directly, e.g. to claim an issue
or mark it "in-progress". `local-issues finish` sets `status = "done"`,
fills in `completed_at`, and records `resolution.notes`.

## CLI commands

Run from anywhere inside the project; `local-issues` walks up from the
current directory to find `.issues/`, the same way git finds `.git/`.

```sh
# create a new open issue
local-issues create -t "Title" -d "Description" \
  --subtask "step one" --subtask "step two"

# list all issues
local-issues list

# list issues with a given status (open, in-progress, done)
local-issues list --status open

# show the title, description, path, and subtasks for one issue
local-issues get <id>

# mark an issue done, record resolution notes, and move it to done/
local-issues finish <id> --notes "Resolution notes"

# move any issues already marked done into done/
local-issues cleanup

# serve a local web UI for browsing issues
local-issues serve
```

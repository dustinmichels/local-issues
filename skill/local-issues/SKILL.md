---
name: local-issues
description: Track and manage project tasks with the `li` CLI, which stores issues as local TOML files in .issues/. Use when asked to create, list, view, claim, update, or finish an issue/task, to find what to work on next, or to set up local issue tracking for a project.
user-invocable: true
---

# Local Issues (li)

`li` tracks issues as TOML files in a `.issues/` directory at the project
root. It's designed to be read and edited by both humans and coding agents —
prefer the CLI for creating and finishing issues, and feel free to
hand-edit TOML files for smaller status/subtask updates.

`li` finds `.issues/` by walking up from the current directory, the same way
git finds `.git/`, so commands work from anywhere in the project.

## Setup

If the project has no `.issues/` directory yet:

```sh
li init
```

This creates `.issues/`, `.issues/done/`, a `README.md` explaining the
layout, and an example issue (`001.toml`).

## Commands

```sh
# create a new open issue, with optional subtasks
li create -t "Title" -d "Description" \
  --subtask "step one" --subtask "step two"

# list all issues (id, title, description, path)
li list

# list issues with a given status
li list --status open          # open | in-progress | done

# show full details of one issue, including subtasks
li get <id>

# mark an issue done, record resolution notes, and move it to .issues/done/
li finish <id> --notes "Resolution notes"

# move any issues already marked status = "done" into .issues/done/
li cleanup

# serve a local web UI for browsing issues
li serve              # default port 7890
li serve --port 8080
```

## Issue file format

Each issue is `.issues/NNN.toml` (zero-padded, e.g. `003.toml`). Finished
issues live in `.issues/done/NNN.toml`.

```toml
[issue]
id = 3
title = "Make modal better"
status = "in-progress"   # open | in-progress | done
assigned_to = ""
created_at = 2026-06-10T20:18:03-04:00

[details]
description = "This is important to me"

[subtasks]
"This is a subtask" = true
"This is another" = false

[resolution]
notes = ""
```

## Working an issue

1. **Find work**: `li list --status open` to see what's available, or
   `li get <id>` for full detail on one issue.
2. **Claim it**: edit the TOML and set `status = "in-progress"` (and
   `assigned_to` if relevant). Hand-editing is fine — `li` doesn't need to
   be invoked for this.
3. **Track progress**: as subtasks complete, flip their value from `false`
   to `true` in `[subtasks]`.
4. **Finish it**: run `li finish <id> --notes "what changed and why"`. This
   sets `status = "done"`, fills in `completed_at`, records the resolution
   notes, and moves the file to `.issues/done/`.

If an issue's `status` was set to `"done"` by hand-editing instead of via
`li finish`, run `li cleanup` to move it into `.issues/done/`.

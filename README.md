# Local Issues (li)

System for tracking issues locally, for use by humans agents.

More sophisticated than a tasks.md, but still based on editing local files.

## Usage

```sh
# create a .issues directory with a README and an example issue
li init

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
li serve --port 8080

# serve the bundled example issues, to try out the web UI
li demo
```

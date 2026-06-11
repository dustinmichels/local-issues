# Local Issues

System for tracking issues locally, for use by humans agents.

More sophisticated than a tasks.md, but still based on editing local files.

## Usage

```sh
# create a .issues directory with a README and an example issue
local-issues init

# create a new issue
local-issues create \
  -t "Lazy-Loaded Background Modal" \
  -d "Calendly currently opens via its default click handlers. Instead, lazy-load an inline Calendly widget into a hidden modal so it can be shown instantly on click rather than loading the widget on demand." \
  --subtask "Create a hidden modal element on the page and initialize Calendly as an inline widget inside it (lazy-loaded after page load or on first user scroll)" \
  --subtask "Replace default Calendly click handlers with a toggle to show/hide this hidden modal, providing an instant popup experience"

# list all issues
local-issues list

# list issues with a given status (open, in-progress, done)
local-issues list --status open

# show the title, description, path, and subtasks for a single issue
local-issues get 1

# mark an issue as done and move it to .issues/done
local-issues finish 1 --notes "Fixed by doing the thing"

# move any issues already marked done into .issues/done
local-issues cleanup

# serve a local web UI for browsing issues
local-issues serve
local-issues serve --port 8080

# serve the bundled example issues, to try out the web UI
local-issues demo
```

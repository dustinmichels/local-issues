# Local Issues

System for tracking issues locally, for use by humans agents.

More sophisticated than a tasks.md, but still based on editing local files.

## Usage

```sh
# create a new issue
local-issues create \
  -t "Lazy-Loaded Background Modal (Instant)" \
  -d "Calendly currently opens via its default click handlers. Instead, lazy-load an inline Calendly widget into a hidden modal so it can be shown instantly on click rather than loading the widget on demand." \
  --subtask "Create a hidden modal element on the page and initialize Calendly as an inline widget inside it (lazy-loaded after page load or on first user scroll)" \
  --subtask "Replace default Calendly click handlers with a toggle to show/hide this hidden modal, providing an instant popup experience"

# list all issues
local-issues list

# list open issues
local-issues list --open
```

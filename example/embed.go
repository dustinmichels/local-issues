// Package example embeds a set of sample issues used by the "demo" command.
package example

import (
	"embed"
	"io/fs"
)

//go:embed all:.issues
var issuesFS embed.FS

// Issues returns the embedded example issues, rooted at the .issues
// directory.
func Issues() (fs.FS, error) {
	return fs.Sub(issuesFS, ".issues")
}

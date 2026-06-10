// Package web embeds the built frontend (web/dist) for serving by the
// local-issues HTTP server.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Dist returns the embedded frontend build, rooted at its dist directory.
func Dist() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}

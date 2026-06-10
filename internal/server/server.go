// Package server provides the local-issues HTTP server, which serves the
// embedded frontend and a JSON API for listing issues.
package server

import (
	"encoding/json"
	"net/http"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/dustinmichels/local-issues/web"
)

// New builds an http.Handler that serves the embedded frontend and an
// /api/issues endpoint backed by the issues in issuesDir.
func New(issuesDir string) (http.Handler, error) {
	dist, err := web.Dist()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/issues", func(w http.ResponseWriter, r *http.Request) {
		summaries, err := issue.List(issuesDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(summaries); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.Handle("/", http.FileServerFS(dist))

	return mux, nil
}

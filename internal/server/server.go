// Package server provides the local-issues HTTP server, which serves the
// embedded frontend and a JSON API for listing and creating issues.
package server

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/dustinmichels/local-issues/web"
)

// createIssueRequest is the JSON body accepted by POST /api/issues.
type createIssueRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Subtasks    []string `json:"subtasks"`
}

// New builds an http.Handler that serves the embedded frontend and a JSON
// API for listing and creating issues in issuesDir.
func New(issuesDir string) (http.Handler, error) {
	dist, err := web.Dist()
	if err != nil {
		return nil, err
	}

	e := echo.New()

	e.GET("/api/issues", func(c *echo.Context) error {
		summaries, err := issue.List(issuesDir)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, summaries)
	})

	e.POST("/api/issues", func(c *echo.Context) error {
		var req createIssueRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		path, err := issue.Create(issue.NewInput{
			Title:       req.Title,
			Description: req.Description,
			Subtasks:    req.Subtasks,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		summary, err := issue.LoadSummary(path)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusCreated, summary)
	})

	e.StaticFS("/", dist)

	return e, nil
}

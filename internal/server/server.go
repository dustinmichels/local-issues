// Package server provides the local-issues HTTP server, which serves the
// embedded frontend and a JSON API for listing and creating issues.
package server

import (
	"errors"
	"net/http"
	"strconv"

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

// updateStatusRequest is the JSON body accepted by
// PATCH /api/issues/:id/status.
type updateStatusRequest struct {
	Status string `json:"status"`
}

// updateIssueRequest is the JSON body accepted by PATCH /api/issues/:id.
type updateIssueRequest struct {
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Status          string          `json:"status"`
	AssignedTo      string          `json:"assigned_to"`
	Subtasks        []issue.Subtask `json:"subtasks"`
	ResolutionNotes string          `json:"resolution_notes"`
}

// issueError maps an error from the issue package to an *echo.HTTPError
// with an appropriate status code.
func issueError(err error) error {
	if errors.Is(err, issue.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
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

	e.GET("/api/issues/:id", func(c *echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
		}

		detail, err := issue.Get(issuesDir, id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}

		return c.JSON(http.StatusOK, detail)
	})

	e.PATCH("/api/issues/:id/status", func(c *echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
		}

		var req updateStatusRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		summary, err := issue.UpdateStatus(issuesDir, id, req.Status)
		if err != nil {
			return issueError(err)
		}

		return c.JSON(http.StatusOK, summary)
	})

	e.PATCH("/api/issues/:id", func(c *echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
		}

		var req updateIssueRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		detail, err := issue.Update(issuesDir, id, issue.UpdateInput{
			Title:           req.Title,
			Description:     req.Description,
			Status:          req.Status,
			AssignedTo:      req.AssignedTo,
			Subtasks:        req.Subtasks,
			ResolutionNotes: req.ResolutionNotes,
		})
		if err != nil {
			return issueError(err)
		}

		return c.JSON(http.StatusOK, detail)
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

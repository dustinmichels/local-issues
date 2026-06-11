// Package issue creates and renders local-issues TOML issue files.
package issue

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// ErrNotFound is returned by Get, Update, and UpdateStatus when no issue
// with the given id exists.
var ErrNotFound = errors.New("issue not found")

// Issue holds the fields needed to render an issue TOML file.
type Issue struct {
	ID          int
	Title       string
	Status      string
	AssignedTo  string
	CreatedAt   time.Time
	CompletedAt time.Time
	Description string
	Subtasks    []string
}

// NewInput holds the user-supplied fields for creating a new issue.
type NewInput struct {
	Title       string
	Description string
	Subtasks    []string
}

// UpdateInput holds the user-editable fields of an issue, as submitted by
// the web UI's edit modal.
type UpdateInput struct {
	Title           string
	Description     string
	Status          string
	AssignedTo      string
	Subtasks        []Subtask
	ResolutionNotes string
}

// validStatuses are the allowed values for an issue's status field.
var validStatuses = map[string]bool{"open": true, "in-progress": true, "done": true}

// New builds a fresh, open Issue from user input, filling in the fields
// that are set automatically (status, created_at, completed_at).
func New(id int, input NewInput) Issue {
	return Issue{
		ID:          id,
		Title:       strings.TrimSpace(input.Title),
		Status:      "open",
		AssignedTo:  "",
		CreatedAt:   time.Now().Truncate(time.Second),
		Description: strings.TrimSpace(input.Description),
		Subtasks:    input.Subtasks,
	}
}

// document is the on-disk TOML shape of an issue file.
type document struct {
	Issue      issueTable      `toml:"issue"`
	Details    detailsTable    `toml:"details"`
	Subtasks   map[string]bool `toml:"subtasks"`
	Resolution resolutionTable `toml:"resolution"`
}

type issueTable struct {
	ID          int       `toml:"id"`
	Title       string    `toml:"title"`
	Status      string    `toml:"status"`
	AssignedTo  string    `toml:"assigned_to"`
	CreatedAt   time.Time `toml:"created_at"`
	CompletedAt time.Time `toml:"completed_at,omitempty"`
}

type detailsTable struct {
	Description string `toml:"description"`
}

type resolutionTable struct {
	Notes string `toml:"notes"`
}

// Render renders an Issue as TOML using the BurntSushi/toml encoder.
func Render(iss Issue) (string, error) {
	subtasks := make(map[string]bool, len(iss.Subtasks))
	for _, subtask := range iss.Subtasks {
		subtasks[subtask] = false
	}

	doc := document{
		Issue: issueTable{
			ID:          iss.ID,
			Title:       iss.Title,
			Status:      iss.Status,
			AssignedTo:  iss.AssignedTo,
			CreatedAt:   iss.CreatedAt,
			CompletedAt: iss.CompletedAt,
		},
		Details:    detailsTable{Description: iss.Description},
		Subtasks:   subtasks,
		Resolution: resolutionTable{Notes: ""},
	}

	var b strings.Builder
	enc := toml.NewEncoder(&b)
	enc.Indent = ""
	if err := enc.Encode(doc); err != nil {
		return "", fmt.Errorf("encoding issue: %w", err)
	}
	return b.String(), nil
}

// Summary holds the fields of an issue needed to render it as a card.
type Summary struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	Path        string     `json:"path"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// summaryDoc is a lenient view of the [issue]/[details] tables used by
// LoadSummary. created_at and completed_at are decoded as `any` so that a
// hand- or agent-edited file with a malformed value (e.g. a stray
// completed_at = "" on an open issue) doesn't fail the whole decode.
type summaryDoc struct {
	Issue struct {
		ID          int    `toml:"id"`
		Title       string `toml:"title"`
		Status      string `toml:"status"`
		CreatedAt   any    `toml:"created_at"`
		CompletedAt any    `toml:"completed_at"`
	} `toml:"issue"`
	Details detailsTable `toml:"details"`
}

// LoadSummary reads the [issue] and [details] tables of the TOML file at
// path and returns its id, title, status, description, path, created_at,
// and (if set) completed_at.
func LoadSummary(path string) (Summary, error) {
	var doc summaryDoc
	if _, err := toml.DecodeFile(path, &doc); err != nil {
		return Summary{}, fmt.Errorf("decoding %s: %w", path, err)
	}

	createdAt, _ := doc.Issue.CreatedAt.(time.Time)

	var completedAt *time.Time
	if t, ok := doc.Issue.CompletedAt.(time.Time); ok {
		completedAt = &t
	}

	return Summary{
		ID:          doc.Issue.ID,
		Title:       doc.Issue.Title,
		Status:      doc.Issue.Status,
		Description: doc.Details.Description,
		Path:        path,
		CreatedAt:   createdAt,
		CompletedAt: completedAt,
	}, nil
}

// Subtask holds a single subtask's text and completion status.
type Subtask struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// Detail holds the full set of fields for a single issue, including its
// subtasks.
type Detail struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	AssignedTo      string    `json:"assigned_to"`
	Description     string    `json:"description"`
	Path            string    `json:"path"`
	Subtasks        []Subtask `json:"subtasks"`
	ResolutionNotes string    `json:"resolution_notes"`
}

// LoadDetail reads the full TOML file at path and returns its id, title,
// description, path, and subtasks.
func LoadDetail(path string) (Detail, error) {
	var doc document
	if _, err := toml.DecodeFile(path, &doc); err != nil {
		return Detail{}, fmt.Errorf("decoding %s: %w", path, err)
	}

	subtasks := make([]Subtask, 0, len(doc.Subtasks))
	for text, done := range doc.Subtasks {
		subtasks = append(subtasks, Subtask{Text: text, Done: done})
	}
	sort.Slice(subtasks, func(i, j int) bool { return subtasks[i].Text < subtasks[j].Text })

	return Detail{
		ID:              doc.Issue.ID,
		Title:           doc.Issue.Title,
		Status:          doc.Issue.Status,
		AssignedTo:      doc.Issue.AssignedTo,
		Description:     doc.Details.Description,
		Path:            path,
		Subtasks:        subtasks,
		ResolutionNotes: doc.Resolution.Notes,
	}, nil
}

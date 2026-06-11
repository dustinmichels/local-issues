// Package issue creates and renders local-issues TOML issue files.
package issue

import (
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

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
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

// LoadSummary reads the [issue] and [details] tables of the TOML file at
// path and returns its id, title, status, description, and path.
func LoadSummary(path string) (Summary, error) {
	var doc map[string]any
	if _, err := toml.DecodeFile(path, &doc); err != nil {
		return Summary{}, fmt.Errorf("decoding %s: %w", path, err)
	}

	issueSection, _ := doc["issue"].(map[string]any)
	detailsSection, _ := doc["details"].(map[string]any)

	id, _ := issueSection["id"].(int64)
	title, _ := issueSection["title"].(string)
	status, _ := issueSection["status"].(string)
	description, _ := detailsSection["description"].(string)

	return Summary{
		ID:          int(id),
		Title:       title,
		Status:      status,
		Description: description,
		Path:        path,
	}, nil
}

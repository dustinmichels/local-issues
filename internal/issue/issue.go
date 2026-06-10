// Package issue creates and renders local-issues TOML issue files.
package issue

import (
	"fmt"
	"regexp"
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
	CreatedAt   string
	CompletedAt string
	Description string
	Tasks       []string
}

// NewInput holds the user-supplied fields for creating a new issue.
type NewInput struct {
	Title       string
	Description string
	Tasks       []string
}

// New builds a fresh, open Issue from user input, filling in the fields
// that are set automatically (status, created_at, completed_at).
func New(id int, input NewInput) Issue {
	return Issue{
		ID:          id,
		Title:       strings.TrimSpace(input.Title),
		Status:      "open",
		AssignedTo:  "",
		CreatedAt:   time.Now().Format("2006-01-02"),
		CompletedAt: "",
		Description: strings.TrimSpace(input.Description),
		Tasks:       input.Tasks,
	}
}

// Render renders an Issue as TOML matching the layout of the example
// template (example/.issues/001.toml).
func Render(iss Issue) string {
	var b strings.Builder

	b.WriteString("[issue]\n")
	fmt.Fprintf(&b, "id = %d\n", iss.ID)
	fmt.Fprintf(&b, "title = %s\n", tomlString(iss.Title))
	fmt.Fprintf(&b, "status = %s # [open, in-progress, done]\n", tomlString(iss.Status))
	fmt.Fprintf(&b, "assigned_to = %s # [claude, antigravity, \"\"]\n", tomlString(iss.AssignedTo))
	fmt.Fprintf(&b, "created_at = %s\n", iss.CreatedAt)
	fmt.Fprintf(&b, "completed_at = %s\n", tomlString(iss.CompletedAt))
	b.WriteString("\n")

	b.WriteString("[details]\n")
	fmt.Fprintf(&b, "description = \"\"\"\n%s\n\"\"\"\n", tomlMultiline(iss.Description))
	b.WriteString("\n")

	b.WriteString("[tasks]\n")
	for _, task := range iss.Tasks {
		fmt.Fprintf(&b, "%s = false\n", tomlString(task))
	}
	b.WriteString("\n")

	b.WriteString("[resolution]\n")
	b.WriteString("notes = \"\"\n")

	return b.String()
}

// tomlString renders s as a TOML basic (double-quoted) string.
func tomlString(s string) string {
	return fmt.Sprintf("%q", s)
}

// tripleQuoteRun matches runs of three or more consecutive double quotes,
// which would otherwise be ambiguous inside a TOML multi-line basic string.
var tripleQuoteRun = regexp.MustCompile(`"{3,}`)

// tomlMultiline escapes s for use inside a TOML multi-line basic string
// ("""..."""), escaping backslashes and any runs of 3+ double quotes.
func tomlMultiline(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return tripleQuoteRun.ReplaceAllStringFunc(s, func(m string) string {
		return strings.Repeat(`\"`, len(m))
	})
}

// Summary holds the fields of an issue needed to render it as a card.
type Summary struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// LoadSummary reads the [issue] table of the TOML file at path and returns
// its id, title, and status.
func LoadSummary(path string) (Summary, error) {
	var doc map[string]any
	if _, err := toml.DecodeFile(path, &doc); err != nil {
		return Summary{}, fmt.Errorf("decoding %s: %w", path, err)
	}

	section, _ := doc["issue"].(map[string]any)

	id, _ := section["id"].(int64)
	title, _ := section["title"].(string)
	status, _ := section["status"].(string)

	return Summary{ID: int(id), Title: title, Status: status}, nil
}

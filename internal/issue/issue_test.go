package issue

import (
	"strings"
	"testing"
	"time"
)

func TestRender(t *testing.T) {
	iss := New(16, NewInput{
		Title:       "Example issue",
		Description: "Example issue. Do not delete.",
		Subtasks:    []string{"subtask 1", "subtask 2", "subtask 3"},
	})

	got, err := Render(iss)
	if err != nil {
		t.Fatal(err)
	}

	want := `[issue]
id = 16
title = "Example issue"
status = "open"
assigned_to = ""
created_at = ` + iss.CreatedAt.Format(time.RFC3339) + `

[details]
description = "Example issue. Do not delete."

[subtasks]
"subtask 1" = false
"subtask 2" = false
"subtask 3" = false

[resolution]
notes = ""
`

	if got != want {
		t.Errorf("Render() mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderNoSubtasks(t *testing.T) {
	iss := New(1, NewInput{Title: "No subtasks", Description: "desc"})
	got, err := Render(iss)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "[subtasks]\n\n[resolution]") {
		t.Errorf("expected empty [subtasks] section, got:\n%s", got)
	}
}

func TestRenderEscaping(t *testing.T) {
	iss := New(1, NewInput{
		Title:       `Fix "quoted" path\name`,
		Description: "Has a literal \"\"\" sequence and a \\ backslash.",
		Subtasks:    []string{`"task" with \backslash`},
	})

	got, err := Render(iss)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, `title = "Fix \"quoted\" path\\name"`) {
		t.Errorf("title not escaped correctly:\n%s", got)
	}
	if !strings.Contains(got, `description = "Has a literal \"\"\" sequence and a \\ backslash."`) {
		t.Errorf("description not escaped correctly:\n%s", got)
	}
	if !strings.Contains(got, `"\"task\" with \\backslash" = false`) {
		t.Errorf("subtask not escaped correctly:\n%s", got)
	}
}

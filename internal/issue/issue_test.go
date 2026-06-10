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
		Tasks:       []string{"subtask 1", "subtask 2", "subtask 3"},
	})

	got := Render(iss)
	want := `[issue]
id = 16
title = "Example issue"
status = "open" # [open, in-progress, done]
assigned_to = "" # [claude, antigravity, ""]
created_at = ` + time.Now().Format("2006-01-02") + `
completed_at = ""

[details]
description = """
Example issue. Do not delete.
"""

[tasks]
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

func TestRenderNoTasks(t *testing.T) {
	iss := New(1, NewInput{Title: "No subtasks", Description: "desc"})
	got := Render(iss)

	if !strings.Contains(got, "[tasks]\n\n[resolution]") {
		t.Errorf("expected empty [tasks] section, got:\n%s", got)
	}
}

func TestRenderEscaping(t *testing.T) {
	iss := New(1, NewInput{
		Title:       `Fix "quoted" path\name`,
		Description: "Has a literal \"\"\" sequence and a \\ backslash.",
		Tasks:       []string{`"task" with \backslash`},
	})

	got := Render(iss)

	if !strings.Contains(got, `title = "Fix \"quoted\" path\\name"`) {
		t.Errorf("title not escaped correctly:\n%s", got)
	}
	if !strings.Contains(got, `Has a literal \"\"\" sequence and a \\ backslash.`) {
		t.Errorf("description not escaped correctly:\n%s", got)
	}
	if !strings.Contains(got, `"\"task\" with \\backslash" = false`) {
		t.Errorf("task not escaped correctly:\n%s", got)
	}
}

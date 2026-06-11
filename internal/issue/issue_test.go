package issue

import (
	"os"
	"reflect"
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

func TestLoadDetail(t *testing.T) {
	dir := t.TempDir()

	iss := New(3, NewInput{
		Title:       "Example issue",
		Description: "Example description",
		Subtasks:    []string{"step b", "step a"},
	})
	data, err := Render(iss)
	if err != nil {
		t.Fatal(err)
	}

	path := dir + "/003.toml"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadDetail(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != 3 || got.Title != "Example issue" || got.Description != "Example description" || got.Path != path {
		t.Errorf("LoadDetail() = %+v, unexpected fields", got)
	}

	want := []Subtask{{Text: "step a", Done: false}, {Text: "step b", Done: false}}
	if !reflect.DeepEqual(got.Subtasks, want) {
		t.Errorf("LoadDetail() Subtasks = %+v, want %+v", got.Subtasks, want)
	}
}

func TestLoadSummary(t *testing.T) {
	dir := t.TempDir()

	iss := New(4, NewInput{Title: "Example issue", Description: "Example description"})
	data, err := Render(iss)
	if err != nil {
		t.Fatal(err)
	}

	path := dir + "/004.toml"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSummary(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != 4 || got.Title != "Example issue" || got.Status != "open" || got.Description != "Example description" || got.Path != path {
		t.Errorf("LoadSummary() = %+v, unexpected fields", got)
	}
	if !got.CreatedAt.Equal(iss.CreatedAt) {
		t.Errorf("LoadSummary() CreatedAt = %v, want %v", got.CreatedAt, iss.CreatedAt)
	}
	if got.CompletedAt != nil {
		t.Errorf("LoadSummary() CompletedAt = %v, want nil", got.CompletedAt)
	}
}

// TestLoadSummaryToleratesEmptyCompletedAt covers files (often hand- or
// agent-edited) that contain a stray completed_at = "" on an open issue,
// which BurntSushi/toml cannot decode directly into a time.Time.
func TestLoadSummaryToleratesEmptyCompletedAt(t *testing.T) {
	dir := t.TempDir()

	data := `[issue]
id = 5
title = "Example issue"
status = "open"
created_at = 2026-06-10
completed_at = ""

[details]
description = "desc"

[subtasks]

[resolution]
notes = ""
`
	path := dir + "/005.toml"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSummary(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != 5 || got.Title != "Example issue" || got.Status != "open" {
		t.Errorf("LoadSummary() = %+v, unexpected fields", got)
	}
	if got.CompletedAt != nil {
		t.Errorf("LoadSummary() CompletedAt = %v, want nil", got.CompletedAt)
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

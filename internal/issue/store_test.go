package issue

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNextID(t *testing.T) {
	dir := t.TempDir()
	doneDir := filepath.Join(dir, DoneDirName)
	if err := os.Mkdir(doneDir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"001.toml", "009.toml", "014.toml", "015.toml"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"002.toml", "013.toml"} {
		if err := os.WriteFile(filepath.Join(doneDir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := NextID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != 16 {
		t.Errorf("NextID() = %d, want 16", got)
	}
}

func TestNextIDEmptyDir(t *testing.T) {
	dir := t.TempDir()

	got, err := NextID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Errorf("NextID() = %d, want 1", got)
	}
}

func TestFindDirWalksUpToAncestor(t *testing.T) {
	root := t.TempDir()
	issuesDir := filepath.Join(root, DirName)
	if err := os.Mkdir(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	restore := chdir(t, nested)
	defer restore()

	got, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	want, err := filepath.EvalSymlinks(issuesDir)
	if err != nil {
		t.Fatal(err)
	}
	gotResolved, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatal(err)
	}
	if gotResolved != want {
		t.Errorf("FindDir() = %s, want %s", gotResolved, want)
	}
}

func TestFindDirDefaultsToCwd(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	restore := chdir(t, nested)
	defer restore()

	got, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	resolvedNested, err := filepath.EvalSymlinks(nested)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(resolvedNested, DirName)
	if got != want {
		t.Errorf("FindDir() = %s, want %s", got, want)
	}
}

func TestInit(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	path, err := Init()
	if err != nil {
		t.Fatal(err)
	}

	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(resolvedRoot, DirName)
	if path != wantPath {
		t.Errorf("Init() path = %s, want %s", path, wantPath)
	}

	if info, err := os.Stat(filepath.Join(path, DoneDirName)); err != nil || !info.IsDir() {
		t.Errorf("expected %s to be a directory", filepath.Join(path, DoneDirName))
	}

	readme, err := os.ReadFile(filepath.Join(path, "README.md"))
	if err != nil {
		t.Fatalf("expected README.md to be created: %v", err)
	}
	if !strings.Contains(string(readme), "local-issues") {
		t.Errorf("README.md missing expected content:\n%s", readme)
	}

	detail, err := Get(path, 1)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Title != "Example issue" || detail.Description != "Example issue. Do not delete.\n" {
		t.Errorf("Get() = %+v, unexpected fields", detail)
	}

	wantSubtasks := []string{"subtask 1", "subtask 2", "subtask 3"}
	if len(detail.Subtasks) != len(wantSubtasks) {
		t.Fatalf("Get() Subtasks = %+v, want %d entries", detail.Subtasks, len(wantSubtasks))
	}
	for i, want := range wantSubtasks {
		if detail.Subtasks[i].Text != want || detail.Subtasks[i].Done {
			t.Errorf("Get() Subtasks[%d] = %+v, want {%s false}", i, detail.Subtasks[i], want)
		}
	}

	summaries, err := List(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0].Status != "open" {
		t.Errorf("List() = %+v, want one open issue", summaries)
	}
}

func TestInitAlreadyExists(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Init(); err != nil {
		t.Fatal(err)
	}

	if _, err := Init(); err == nil {
		t.Error("expected error for already-initialized directory, got nil")
	}
}

func TestCreate(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	path, err := Create(NewInput{
		Title:       "First issue",
		Description: "Something to do",
		Subtasks:    []string{"step one"},
	})
	if err != nil {
		t.Fatal(err)
	}

	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(resolvedRoot, DirName, "001.toml")
	if path != wantPath {
		t.Errorf("Create() path = %s, want %s", path, wantPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "id = 1\n") || !strings.Contains(string(data), `title = "First issue"`) {
		t.Errorf("unexpected file contents:\n%s", data)
	}
}

func TestFinish(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{
		Title:       "Fix bug",
		Description: "Something is broken",
		Subtasks:    []string{"step one"},
	}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	path, err := Finish(issuesDir, 1, "  Fixed by doing the thing.  ")
	if err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(issuesDir, DoneDirName, "001.toml")
	if path != wantPath {
		t.Errorf("Finish() path = %s, want %s", path, wantPath)
	}

	if _, err := os.Stat(filepath.Join(issuesDir, "001.toml")); !os.IsNotExist(err) {
		t.Errorf("expected original file to be removed, stat err = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	if !strings.Contains(got, `status = "done"`) {
		t.Errorf("expected status = \"done\":\n%s", got)
	}
	if !strings.Contains(got, "completed_at = ") {
		t.Errorf("expected completed_at to be set:\n%s", got)
	}
	if !strings.Contains(got, `notes = "Fixed by doing the thing."`) {
		t.Errorf("expected trimmed resolution notes:\n%s", got)
	}
	if !strings.Contains(got, `"step one" = false`) {
		t.Errorf("expected subtasks to be preserved:\n%s", got)
	}
}

func TestFinishMissingIssue(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Finish(issuesDir, 1, "notes"); err == nil {
		t.Error("expected error for missing issue, got nil")
	}
}

func TestCleanup(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Open issue", Description: "desc"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(NewInput{Title: "Done issue", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	// Mark issue 2 as done without moving it, as if it were finished by hand.
	donePath := filepath.Join(issuesDir, FileName(2))
	data, err := os.ReadFile(donePath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(data), `status = "open"`, `status = "done"`, 1)
	if err := os.WriteFile(donePath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	moved, err := Cleanup(issuesDir)
	if err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(issuesDir, DoneDirName, FileName(2))
	if len(moved) != 1 || moved[0] != wantPath {
		t.Errorf("Cleanup() = %v, want [%s]", moved, wantPath)
	}

	if _, err := os.Stat(donePath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, stat err = %v", donePath, err)
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Errorf("expected %s to exist: %v", wantPath, err)
	}

	openPath := filepath.Join(issuesDir, FileName(1))
	if _, err := os.Stat(openPath); err != nil {
		t.Errorf("expected open issue to remain at %s: %v", openPath, err)
	}
}

func TestCleanupNoneDone(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Open issue", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	moved, err := Cleanup(issuesDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(moved) != 0 {
		t.Errorf("Cleanup() = %v, want none", moved)
	}
}

func TestCleanupMissingDir(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	moved, err := Cleanup(issuesDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(moved) != 0 {
		t.Errorf("Cleanup() = %v, want none", moved)
	}
}

func TestListIncludesUnfiledDoneIssue(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Open issue", Description: "desc"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(NewInput{Title: "Done issue", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	// Mark issue 2 as done without moving it, as if it were finished by hand.
	path := filepath.Join(issuesDir, FileName(2))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(data), `status = "open"`, `status = "done"`, 1)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	summaries, err := List(issuesDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(summaries) != 2 {
		t.Fatalf("List() = %+v, want 2 summaries", summaries)
	}
	if summaries[1].ID != 2 || summaries[1].Status != "done" || summaries[1].Path != path {
		t.Errorf("List()[1] = %+v, want id 2, status done, path %s", summaries[1], path)
	}
}

func TestGet(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{
		Title:       "Fix bug",
		Description: "Something is broken",
		Subtasks:    []string{"step one"},
	}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	got, err := Get(issuesDir, 1)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != 1 || got.Title != "Fix bug" || got.Description != "Something is broken" {
		t.Errorf("Get() = %+v, unexpected fields", got)
	}
	if len(got.Subtasks) != 1 || got.Subtasks[0].Text != "step one" || got.Subtasks[0].Done {
		t.Errorf("Get() Subtasks = %+v, want [{step one false}]", got.Subtasks)
	}
}

func TestGetFindsDoneIssue(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Fix bug", Description: "Something is broken"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Finish(issuesDir, 1, "done"); err != nil {
		t.Fatal(err)
	}

	got, err := Get(issuesDir, 1)
	if err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(issuesDir, DoneDirName, FileName(1))
	if got.Path != wantPath {
		t.Errorf("Get() Path = %s, want %s", got.Path, wantPath)
	}
}

func TestGetMissingIssue(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Get(issuesDir, 1); err == nil {
		t.Error("expected error for missing issue, got nil")
	}
}

func TestCreateRequiresTitle(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "  ", Description: "desc"}); err == nil {
		t.Error("expected error for blank title, got nil")
	}
}

func TestDelete(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Fix bug", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(issuesDir, FileName(1))
	if err := Delete(issuesDir, 1); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, stat err = %v", path, err)
	}
}

func TestDeleteFromDoneDir(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Fix bug", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	path, err := Finish(issuesDir, 1, "done")
	if err != nil {
		t.Fatal(err)
	}

	if err := Delete(issuesDir, 1); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s to be removed, stat err = %v", path, err)
	}
}

func TestDeleteMissingIssue(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := Delete(issuesDir, 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete() err = %v, want ErrNotFound", err)
	}
}

func TestUpdateStatus(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Fix bug", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	summary, err := UpdateStatus(issuesDir, 1, "in-progress")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Status != "in-progress" {
		t.Errorf("UpdateStatus() Status = %q, want %q", summary.Status, "in-progress")
	}
	if summary.CompletedAt != nil {
		t.Errorf("UpdateStatus() CompletedAt = %v, want nil", summary.CompletedAt)
	}

	path := filepath.Join(issuesDir, FileName(1))
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected issue to remain at %s: %v", path, err)
	}

	summary, err = UpdateStatus(issuesDir, 1, "done")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Status != "done" {
		t.Errorf("UpdateStatus() Status = %q, want %q", summary.Status, "done")
	}
	if summary.CompletedAt == nil {
		t.Error("UpdateStatus() CompletedAt = nil, want non-nil")
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected issue to remain at %s without being moved: %v", path, err)
	}
}

func TestUpdateStatusFromDoneClearsCompletedAt(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Fix bug", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Finish(issuesDir, 1, "done"); err != nil {
		t.Fatal(err)
	}

	summary, err := UpdateStatus(issuesDir, 1, "open")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Status != "open" {
		t.Errorf("UpdateStatus() Status = %q, want %q", summary.Status, "open")
	}
	if summary.CompletedAt != nil {
		t.Errorf("UpdateStatus() CompletedAt = %v, want nil", summary.CompletedAt)
	}

	path := filepath.Join(issuesDir, DoneDirName, FileName(1))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "completed_at") {
		t.Errorf("expected completed_at to be cleared:\n%s", data)
	}
}

func TestUpdateStatusInvalid(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Fix bug", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := UpdateStatus(issuesDir, 1, "bogus"); err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestUpdateStatusMissingIssue(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := UpdateStatus(issuesDir, 1, "open"); err == nil {
		t.Error("expected error for missing issue, got nil")
	}
}

func TestUpdate(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{
		Title:       "Fix bug",
		Description: "Something is broken",
		Subtasks:    []string{"step one"},
	}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	detail, err := Update(issuesDir, 1, UpdateInput{
		Title:       "Fix the bug",
		Description: "Updated description",
		Status:      "in-progress",
		AssignedTo:  "dustin",
		Subtasks: []Subtask{
			{Text: "step one", Done: true},
			{Text: "step two", Done: false},
		},
		ResolutionNotes: "",
	})
	if err != nil {
		t.Fatal(err)
	}

	if detail.Title != "Fix the bug" || detail.Description != "Updated description" ||
		detail.Status != "in-progress" || detail.AssignedTo != "dustin" {
		t.Errorf("Update() = %+v, unexpected fields", detail)
	}

	wantSubtasks := map[string]bool{"step one": true, "step two": false}
	if len(detail.Subtasks) != len(wantSubtasks) {
		t.Fatalf("Update() Subtasks = %+v, want %d entries", detail.Subtasks, len(wantSubtasks))
	}
	for _, s := range detail.Subtasks {
		if want, ok := wantSubtasks[s.Text]; !ok || want != s.Done {
			t.Errorf("Update() Subtask %+v, want done=%v", s, want)
		}
	}

	path := filepath.Join(issuesDir, FileName(1))
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected issue to remain at %s: %v", path, err)
	}

	// Marking the issue done via Update should set completed_at, even
	// though the file isn't moved into done/.
	detail, err = Update(issuesDir, 1, UpdateInput{
		Title:           "Fix the bug",
		Description:     "Updated description",
		Status:          "done",
		AssignedTo:      "dustin",
		Subtasks:        []Subtask{{Text: "step one", Done: true}},
		ResolutionNotes: "Fixed it",
	})
	if err != nil {
		t.Fatal(err)
	}
	if detail.Status != "done" || detail.ResolutionNotes != "Fixed it" {
		t.Errorf("Update() = %+v, unexpected fields", detail)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "completed_at = ") {
		t.Errorf("expected completed_at to be set:\n%s", data)
	}
}

func TestUpdateRequiresTitle(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Fix bug", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Update(issuesDir, 1, UpdateInput{Title: "  ", Status: "open"}); err == nil {
		t.Error("expected error for blank title, got nil")
	}
}

func TestUpdateInvalidStatus(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "Fix bug", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Update(issuesDir, 1, UpdateInput{Title: "Fix bug", Status: "bogus"}); err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestUpdateMissingIssue(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	issuesDir, err := FindDir()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Update(issuesDir, 1, UpdateInput{Title: "Fix bug", Status: "open"}); err == nil {
		t.Error("expected error for missing issue, got nil")
	}
}

// chdir changes the working directory to dir for the duration of the test
// and returns a function that restores the original directory.
func chdir(t *testing.T, dir string) func() {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Chdir(old); err != nil {
			t.Fatal(err)
		}
	}
}

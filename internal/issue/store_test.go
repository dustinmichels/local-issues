package issue

import (
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

func TestCreate(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	path, err := Create(NewInput{
		Title:       "First issue",
		Description: "Something to do",
		Tasks:       []string{"step one"},
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

func TestCreateRequiresTitle(t *testing.T) {
	root := t.TempDir()
	restore := chdir(t, root)
	defer restore()

	if _, err := Create(NewInput{Title: "  ", Description: "desc"}); err == nil {
		t.Error("expected error for blank title, got nil")
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

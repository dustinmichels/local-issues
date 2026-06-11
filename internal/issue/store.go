package issue

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	// DirName is the name of the directory that holds issue files.
	DirName = ".issues"
	// DoneDirName is the subdirectory of DirName that holds completed issues.
	DoneDirName = "done"
)

// idFilePattern matches issue filenames such as "001.toml".
var idFilePattern = regexp.MustCompile(`^(\d+)\.toml$`)

// FindDir locates the .issues directory by checking the current directory
// and walking up through its ancestors, the same way git locates .git. If
// no .issues directory is found, it returns the path for one in the current
// directory without creating it.
func FindDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for dir := cwd; ; {
		candidate := filepath.Join(dir, DirName)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Join(cwd, DirName), nil
		}
		dir = parent
	}
}

// NextID returns the next available issue ID by scanning issuesDir and its
// done subdirectory for the highest existing "<id>.toml" file.
func NextID(issuesDir string) (int, error) {
	maxID := 0

	for _, dir := range []string{issuesDir, filepath.Join(issuesDir, DoneDirName)} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return 0, fmt.Errorf("reading %s: %w", dir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			match := idFilePattern.FindStringSubmatch(entry.Name())
			if match == nil {
				continue
			}
			id, err := strconv.Atoi(match[1])
			if err != nil {
				continue
			}
			if id > maxID {
				maxID = id
			}
		}
	}

	return maxID + 1, nil
}

// FileName returns the issue filename for the given ID, e.g. "007.toml".
func FileName(id int) string {
	return fmt.Sprintf("%03d.toml", id)
}

// exampleIssueTemplate is the content of the example issue written by
// Init. The %s placeholder is filled with today's date in YYYY-MM-DD form.
const exampleIssueTemplate = `[issue]
id = 1
title = "Example issue"
status = "open" # [open, in-progress, done]
assigned_to = ""
created_at = %s

[details]
description = """
Example issue. Do not delete.
"""

[subtasks]
"subtask 1" = false
"subtask 2" = false
"subtask 3" = false

[resolution]
notes = ""
`

// issuesReadme is the content of the README.md written by Init,
// documenting the .issues directory layout and the local-issues CLI.
//
//go:embed issues_readme.md
var issuesReadme string

// Init creates a new .issues directory with its done subdirectory, a
// README documenting the CLI, and an example issue, and returns the path
// of the created directory. It returns an error if the directory already
// exists.
func Init() (string, error) {
	issuesDir, err := FindDir()
	if err != nil {
		return "", err
	}

	if info, err := os.Stat(issuesDir); err == nil && info.IsDir() {
		return "", fmt.Errorf("issues directory already exists: %s", issuesDir)
	}

	doneDir := filepath.Join(issuesDir, DoneDirName)
	if err := os.MkdirAll(doneDir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", doneDir, err)
	}

	readmePath := filepath.Join(issuesDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(issuesReadme), 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", readmePath, err)
	}

	examplePath := filepath.Join(issuesDir, FileName(1))
	data := fmt.Sprintf(exampleIssueTemplate, time.Now().Format("2006-01-02"))
	if err := os.WriteFile(examplePath, []byte(data), 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", examplePath, err)
	}

	return issuesDir, nil
}

// Create writes a new open issue to the .issues directory and returns the
// path of the created file.
func Create(input NewInput) (string, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return "", fmt.Errorf("title is required")
	}

	issuesDir, err := FindDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", issuesDir, err)
	}

	id, err := NextID(issuesDir)
	if err != nil {
		return "", err
	}

	iss := New(id, input)
	path := filepath.Join(issuesDir, FileName(id))

	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("issue file already exists: %s", path)
	}

	data, err := Render(iss)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}

	return path, nil
}

// Get loads the full detail of the issue with the given id, checking
// issuesDir and its done subdirectory for "<id>.toml". It returns an
// error if no issue with that id exists.
func Get(issuesDir string, id int) (Detail, error) {
	for _, dir := range []string{issuesDir, filepath.Join(issuesDir, DoneDirName)} {
		path := filepath.Join(dir, FileName(id))
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Detail{}, fmt.Errorf("checking %s: %w", path, err)
		}
		return LoadDetail(path)
	}

	return Detail{}, fmt.Errorf("issue %d not found", id)
}

// Finish marks the issue with the given id as done, records the
// resolution notes and completion time, and moves its file into the
// done subdirectory of issuesDir. It returns the path of the moved file.
func Finish(issuesDir string, id int, notes string) (string, error) {
	srcPath := filepath.Join(issuesDir, FileName(id))

	if _, err := os.Stat(srcPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("issue %d not found", id)
		}
		return "", fmt.Errorf("checking %s: %w", srcPath, err)
	}

	var doc document
	if _, err := toml.DecodeFile(srcPath, &doc); err != nil {
		return "", fmt.Errorf("decoding %s: %w", srcPath, err)
	}

	doc.Issue.Status = "done"
	doc.Issue.CompletedAt = time.Now().Truncate(time.Second)
	doc.Resolution.Notes = strings.TrimSpace(notes)

	var b strings.Builder
	enc := toml.NewEncoder(&b)
	enc.Indent = ""
	if err := enc.Encode(doc); err != nil {
		return "", fmt.Errorf("encoding issue %d: %w", id, err)
	}

	doneDir := filepath.Join(issuesDir, DoneDirName)
	if err := os.MkdirAll(doneDir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", doneDir, err)
	}

	destPath := filepath.Join(doneDir, FileName(id))
	if err := os.WriteFile(destPath, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", destPath, err)
	}

	if err := os.Remove(srcPath); err != nil {
		return "", fmt.Errorf("removing %s: %w", srcPath, err)
	}

	return destPath, nil
}

// Cleanup moves any issue files directly in issuesDir whose status is
// "done" into the done subdirectory. It returns the destination paths of
// the moved files, sorted by ID ascending.
func Cleanup(issuesDir string) ([]string, error) {
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", issuesDir, err)
	}

	doneDir := filepath.Join(issuesDir, DoneDirName)

	var moved []string
	for _, entry := range entries {
		if entry.IsDir() || !idFilePattern.MatchString(entry.Name()) {
			continue
		}

		srcPath := filepath.Join(issuesDir, entry.Name())
		summary, err := LoadSummary(srcPath)
		if err != nil {
			return nil, err
		}
		if summary.Status != "done" {
			continue
		}

		if err := os.MkdirAll(doneDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating %s: %w", doneDir, err)
		}

		destPath := filepath.Join(doneDir, entry.Name())
		if err := os.Rename(srcPath, destPath); err != nil {
			return nil, fmt.Errorf("moving %s: %w", srcPath, err)
		}

		moved = append(moved, destPath)
	}

	sort.Strings(moved)

	return moved, nil
}

// List returns a summary of every issue in issuesDir and its done
// subdirectory, sorted by ID ascending.
func List(issuesDir string) ([]Summary, error) {
	var summaries []Summary

	for _, dir := range []string{issuesDir, filepath.Join(issuesDir, DoneDirName)} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading %s: %w", dir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !idFilePattern.MatchString(entry.Name()) {
				continue
			}

			summary, err := LoadSummary(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, err
			}
			summaries = append(summaries, summary)
		}
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].ID < summaries[j].ID
	})

	return summaries, nil
}

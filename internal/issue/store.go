package issue

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

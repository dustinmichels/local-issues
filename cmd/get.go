package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show the title, description, path, and subtasks of an issue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid issue id %q", args[0])
		}

		issuesDir, err := issue.FindDir()
		if err != nil {
			return err
		}

		detail, err := issue.Get(issuesDir, id)
		if err != nil {
			return err
		}

		path := detail.Path
		if cwd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(cwd, detail.Path); err == nil {
				path = rel
			}
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Title: %s\n", detail.Title)
		fmt.Fprintf(out, "Description: %s\n", detail.Description)
		fmt.Fprintf(out, "Path: %s\n", path)

		fmt.Fprintln(out, "Subtasks:")
		if len(detail.Subtasks) == 0 {
			fmt.Fprintln(out, "  (none)")
		} else {
			for _, s := range detail.Subtasks {
				mark := " "
				if s.Done {
					mark = "x"
				}
				fmt.Fprintf(out, "  [%s] %s\n", mark, s.Text)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

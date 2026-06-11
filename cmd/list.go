package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/spf13/cobra"
)

var listStatus string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.FindDir()
		if err != nil {
			return err
		}

		summaries, err := issue.List(issuesDir)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		defer w.Flush()

		fmt.Fprintln(w, "ID\tSTATUS\tTITLE\tDESCRIPTION\tPATH")
		for _, s := range summaries {
			if listStatus != "" && s.Status != listStatus {
				continue
			}

			path := s.Path
			if rel, err := filepath.Rel(cwd, s.Path); err == nil {
				path = rel
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", s.ID, s.Status, s.Title, s.Description, path)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listStatus, "status", "", "only list issues with this status (open, in-progress, done)")
}

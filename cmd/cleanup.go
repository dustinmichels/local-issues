package cmd

import (
	"fmt"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Move issues marked done into the done directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.FindDir()
		if err != nil {
			return err
		}

		moved, err := issue.Cleanup(issuesDir)
		if err != nil {
			return err
		}

		if len(moved) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No done issues to move")
			return nil
		}

		for _, path := range moved {
			fmt.Fprintf(cmd.OutOrStdout(), "Moved %s\n", path)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}

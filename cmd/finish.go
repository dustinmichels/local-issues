package cmd

import (
	"fmt"
	"strconv"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/spf13/cobra"
)

var finishNotes string

var finishCmd = &cobra.Command{
	Use:   "finish <id>",
	Short: "Mark an issue as done and move it to the done directory",
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

		path, err := issue.Finish(issuesDir, id, finishNotes)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Finished %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(finishCmd)

	finishCmd.Flags().StringVar(&finishNotes, "notes", "", "resolution notes")
}

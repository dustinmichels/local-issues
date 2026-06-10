package cmd

import (
	"fmt"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/spf13/cobra"
)

var (
	createTitle       string
	createDescription string
	createTasks       []string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new open issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := issue.Create(issue.NewInput{
			Title:       createTitle,
			Description: createDescription,
			Tasks:       createTasks,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "issue title (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "issue description")
	createCmd.Flags().StringArrayVar(&createTasks, "task", nil, "a subtask (repeatable)")

	createCmd.MarkFlagRequired("title")
	createCmd.MarkFlagRequired("description")
}

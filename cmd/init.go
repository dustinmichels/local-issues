package cmd

import (
	"fmt"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new .issues directory with an example issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := issue.Init()
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Initialized %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

package cmd

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/dustinmichels/local-issues/internal/issue"
	"github.com/dustinmichels/local-issues/internal/server"
	"github.com/spf13/cobra"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a local web UI for viewing issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesDir, err := issue.FindDir()
		if err != nil {
			return err
		}

		handler, err := server.New(issuesDir)
		if err != nil {
			return err
		}

		addr := fmt.Sprintf("127.0.0.1:%d", servePort)
		url := fmt.Sprintf("http://%s", addr)

		fmt.Fprintf(cmd.OutOrStdout(), "Serving issues from %s at %s\n", issuesDir, url)
		openBrowser(url)

		return http.ListenAndServe(addr, handler)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntVar(&servePort, "port", 7890, "port to serve on")
}

// openBrowser opens url in the default browser, ignoring any error.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

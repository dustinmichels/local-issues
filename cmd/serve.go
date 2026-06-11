package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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

		srv := &http.Server{Addr: addr, Handler: handler}

		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		serveErr := make(chan error, 1)
		go func() {
			serveErr <- srv.ListenAndServe()
		}()

		select {
		case err := <-serveErr:
			if err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		case <-ctx.Done():
		}

		fmt.Fprintln(cmd.OutOrStdout(), "\nShutting down, moving done issues into done/...")
		moved, err := issue.Cleanup(issuesDir)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "cleanup failed: %v\n", err)
		}
		for _, path := range moved {
			fmt.Fprintf(cmd.OutOrStdout(), "Moved %s\n", path)
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
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

package cmd

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dustinmichels/local-issues/example"
	"github.com/dustinmichels/local-issues/internal/server"
	"github.com/spf13/cobra"
)

var demoPort int

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Serve the bundled example issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		issuesFS, err := example.Issues()
		if err != nil {
			return err
		}

		issuesDir, err := os.MkdirTemp("", "local-issues-demo-*")
		if err != nil {
			return err
		}

		if err := extractFS(issuesFS, issuesDir); err != nil {
			return err
		}

		handler, err := server.New(issuesDir)
		if err != nil {
			return err
		}

		addr := fmt.Sprintf("127.0.0.1:%d", demoPort)
		url := fmt.Sprintf("http://%s", addr)

		fmt.Fprintf(cmd.OutOrStdout(), "Serving example issues at %s\n", url)
		openBrowser(url)

		return http.ListenAndServe(addr, handler)
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)

	demoCmd.Flags().IntVar(&demoPort, "port", 7891, "port to serve on")
}

// extractFS copies every file in src to dst on disk, preserving its
// directory structure.
func extractFS(src fs.FS, dst string) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		target := filepath.Join(dst, path)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

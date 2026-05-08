package cmd

import (
	"fmt"
	"os"

	"github.com/muhammedikinci/pin/internal/runner"
	"github.com/spf13/cobra"
)

var daemonHost string
var daemonPort int
var daemonToken string
var daemonMaxConcurrent int
var daemonDataDir string
var daemonGitHubSecret string
var daemonGitHubBranch string
var daemonGitHubPipeline string

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run Pin as a VPS pipeline daemon",
	Long: `Start the HTTP and Server-Sent Events daemon.

Run this on your VPS, then use "pin trigger" from another machine to send a
pipeline and "pin watch" to follow live events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if daemonPort <= 0 || daemonPort > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
		if daemonMaxConcurrent < 1 {
			return fmt.Errorf("max-concurrent must be at least 1")
		}

		return runner.ApplyDaemonWithOptions("", runner.DaemonOptions{
			Host:               daemonHost,
			Port:               daemonPort,
			Token:              daemonToken,
			MaxConcurrent:      daemonMaxConcurrent,
			DataDir:            daemonDataDir,
			GitHubSecret:       daemonGitHubSecret,
			GitHubBranch:       daemonGitHubBranch,
			GitHubPipelineFile: daemonGitHubPipeline,
		})
	},
}

func init() {
	daemonCmd.Flags().StringVar(&daemonHost, "host", "127.0.0.1", "host address for the daemon to bind")
	daemonCmd.Flags().IntVar(&daemonPort, "port", 8081, "port for the daemon to listen on")
	daemonCmd.Flags().StringVar(&daemonToken, "token", os.Getenv("PIN_TOKEN"), "bearer token for remote access; defaults to PIN_TOKEN")
	daemonCmd.Flags().IntVar(&daemonMaxConcurrent, "max-concurrent", 1, "maximum number of pipelines to run at once")
	daemonCmd.Flags().StringVar(&daemonDataDir, "data-dir", ".pin", "directory for persistent run metadata and logs")
	daemonCmd.Flags().StringVar(&daemonGitHubSecret, "github-secret", os.Getenv("PIN_GITHUB_SECRET"), "GitHub webhook secret; defaults to PIN_GITHUB_SECRET")
	daemonCmd.Flags().StringVar(&daemonGitHubBranch, "github-branch", "main", "GitHub branch allowed to trigger webhook pipelines")
	daemonCmd.Flags().StringVar(&daemonGitHubPipeline, "github-pipeline", "", "pipeline file used by /webhooks/github")
	rootCmd.AddCommand(daemonCmd)
}

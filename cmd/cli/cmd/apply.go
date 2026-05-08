package cmd

import (
	"fmt"

	"github.com/muhammedikinci/pin/internal/runner"
	"github.com/spf13/cobra"
)

var pipelineName string
var pipelineFilePath string
var daemonMode bool

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Run a pipeline, kept as a backward-compatible alias",
	Long: `Run a pipeline from a YAML file.

This command is kept for existing users. Prefer "pin run" for local runs and
"pin daemon" plus "pin trigger" for VPS usage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if daemonMode {
			return runner.ApplyDaemon(pipelineFilePath)
		}
		return runner.Apply(pipelineFilePath)
	},
}

func init() {
	applyCmd.PersistentFlags().StringVarP(&pipelineName, "name", "n", "", "pipeline name")
	applyCmd.PersistentFlags().StringVarP(&pipelineFilePath, "filepath", "f", "", "pipeline configuration file path")
	applyCmd.PersistentFlags().BoolVar(&daemonMode, "daemon", false, "run as daemon with SSE server for real-time event streaming")

	// In daemon mode, filepath is optional since it will be provided via HTTP endpoint later
	applyCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if !daemonMode && pipelineFilePath == "" {
			return fmt.Errorf("required flag \"filepath\" not set")
		}
		return nil
	}

	rootCmd.AddCommand(applyCmd)
}

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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if daemonMode {
			runner.ApplyDaemon(pipelineFilePath)
		} else {
			runner.Apply(pipelineFilePath)
		}
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

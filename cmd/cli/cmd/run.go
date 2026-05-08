package cmd

import (
	"fmt"

	"github.com/muhammedikinci/pin/internal/runner"
	"github.com/spf13/cobra"
)

var runFilePath string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a pipeline locally",
	Long: `Run a Docker-based pipeline from a YAML file on this machine.

Use this while developing locally. For VPS usage, start "pin daemon" on the
server and send pipelines with "pin trigger".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if runFilePath == "" {
			return fmt.Errorf("required flag \"filepath\" not set")
		}

		return runner.Apply(runFilePath)
	},
}

func init() {
	runCmd.Flags().StringVarP(&runFilePath, "filepath", "f", "pin.yaml", "pipeline configuration file path")
	rootCmd.AddCommand(runCmd)
}

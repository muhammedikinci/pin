package cmd

import (
	"github.com/muhammedikinci/pin/internal/runner"
	"github.com/spf13/cobra"
)

var pipelineName string
var pipelineFilePath string

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
		runner.Apply(pipelineFilePath)
	},
}

func init() {
	applyCmd.PersistentFlags().StringVarP(&pipelineName, "name", "n", "", "pipeline name")
	applyCmd.PersistentFlags().StringVarP(&pipelineFilePath, "filepath", "f", "", "pipeline configuration file path")

	applyCmd.MarkPersistentFlagRequired("filepath")

	rootCmd.AddCommand(applyCmd)
}

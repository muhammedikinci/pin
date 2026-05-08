package cmd

import (
	"fmt"

	"github.com/muhammedikinci/pin/internal/runner"
	"github.com/spf13/cobra"
)

var validateFilePath string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a pipeline file without running it",
	Long:  "Validate a pin.yaml file before sending it to a VPS daemon or running it locally.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if validateFilePath == "" {
			return fmt.Errorf("required flag \"filepath\" not set")
		}

		if err := runner.Validate(validateFilePath); err != nil {
			return err
		}

		fmt.Printf("Pipeline configuration is valid: %s\n", validateFilePath)
		return nil
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validateFilePath, "filepath", "f", "pin.yaml", "pipeline configuration file path")
	rootCmd.AddCommand(validateCmd)
}

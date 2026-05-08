package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pin",
	Short: "A tiny VPS-friendly pipeline runner",
	Long: `Pin runs Docker-based build, test, and deploy pipelines without making you
set up a full CI system.

Use it locally while developing, or run pin daemon on your own VPS and trigger
pipelines over HTTP while watching live Server-Sent Events logs from anywhere.`,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}

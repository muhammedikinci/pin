package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var logsURL string
var logsToken string

var logsCmd = &cobra.Command{
	Use:   "logs <run_id>",
	Short: "Fetch logs for a daemon run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint, err := normalizeRemoteEndpoint(logsURL, "/runs/"+args[0]+"/logs")
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		addToken(req, logsToken)

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf("daemon returned %s", resp.Status)
		}

		_, err = io.Copy(os.Stdout, resp.Body)
		return err
	},
}

func init() {
	logsCmd.Flags().StringVar(&logsURL, "url", defaultDaemonURL, "Pin daemon URL")
	logsCmd.Flags().StringVar(&logsToken, "token", os.Getenv("PIN_TOKEN"), "bearer token; defaults to PIN_TOKEN")
	rootCmd.AddCommand(logsCmd)
}

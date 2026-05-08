package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var runsURL string
var runsToken string

var runsCmd = &cobra.Command{
	Use:   "runs",
	Short: "List recent daemon runs",
	Long:  "Fetch recent pipeline run status from a Pin daemon.",
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint, err := normalizeRemoteEndpoint(runsURL, "/runs")
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		addToken(req, runsToken)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf("daemon returned %s", resp.Status)
		}

		var response struct {
			Runs []struct {
				ID          string `json:"id"`
				Status      string `json:"status"`
				Source      string `json:"source"`
				StartedAt   string `json:"started_at"`
				CompletedAt string `json:"completed_at,omitempty"`
				Error       string `json:"error,omitempty"`
			} `json:"runs"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		if len(response.Runs) == 0 {
			fmt.Println("No runs recorded yet.")
			return nil
		}

		fmt.Println("RUN ID\tSTATUS\tSOURCE\tSTARTED\tERROR")
		for _, run := range response.Runs {
			errorText := run.Error
			if errorText == "" {
				errorText = "-"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", run.ID, run.Status, run.Source, run.StartedAt, errorText)
		}
		return nil
	},
}

func init() {
	runsCmd.Flags().StringVar(&runsURL, "url", defaultDaemonURL, "Pin daemon URL")
	runsCmd.Flags().StringVar(&runsToken, "token", os.Getenv("PIN_TOKEN"), "bearer token; defaults to PIN_TOKEN")
	rootCmd.AddCommand(runsCmd)
}

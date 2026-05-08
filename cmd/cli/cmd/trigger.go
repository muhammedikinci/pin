package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var triggerFilePath string
var triggerURL string
var triggerToken string
var triggerTimeout time.Duration

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Send a pipeline file to a Pin daemon",
	Long:  "Trigger a pipeline on a remote Pin daemon by POSTing a YAML file to /trigger.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if triggerFilePath == "" {
			return fmt.Errorf("required flag \"filepath\" not set")
		}

		yamlContent, err := os.ReadFile(triggerFilePath)
		if err != nil {
			return err
		}

		endpoint, err := normalizeRemoteEndpoint(triggerURL, "/trigger")
		if err != nil {
			return err
		}

		client := &http.Client{Timeout: triggerTimeout}
		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(yamlContent))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/yaml")
		addToken(req, triggerToken)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf("daemon returned %s: %s", resp.Status, string(body))
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("Pipeline accepted by %s\n", endpoint)
			return nil
		}

		fmt.Printf("Pipeline accepted by %s\n", endpoint)
		if runID, ok := result["run_id"].(string); ok && runID != "" {
			fmt.Printf("Run ID: %s\n", runID)
		}
		return nil
	},
}

func init() {
	triggerCmd.Flags().StringVarP(&triggerFilePath, "filepath", "f", "pin.yaml", "pipeline configuration file path")
	triggerCmd.Flags().StringVar(&triggerURL, "url", defaultDaemonURL, "Pin daemon URL")
	triggerCmd.Flags().StringVar(&triggerToken, "token", os.Getenv("PIN_TOKEN"), "bearer token; defaults to PIN_TOKEN")
	triggerCmd.Flags().DurationVar(&triggerTimeout, "timeout", 15*time.Second, "HTTP request timeout")
	rootCmd.AddCommand(triggerCmd)
}

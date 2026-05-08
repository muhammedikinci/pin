package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var watchURL string
var watchToken string

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch live events from a Pin daemon",
	Long:  "Connect to a Pin daemon's /events endpoint and print live pipeline events.",
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint, err := normalizeRemoteEndpoint(watchURL, "/events")
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		addToken(req, watchToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf("daemon returned %s", resp.Status)
		}

		fmt.Printf("Watching %s\n", endpoint)

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024), 1024*1024)
		eventType := "message"

		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event:"):
				eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:"):
				printSSEEvent(eventType, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			}
		}

		return scanner.Err()
	},
}

func init() {
	watchCmd.Flags().StringVar(&watchURL, "url", defaultDaemonURL, "Pin daemon URL")
	watchCmd.Flags().StringVar(&watchToken, "token", os.Getenv("PIN_TOKEN"), "bearer token; defaults to PIN_TOKEN")
	rootCmd.AddCommand(watchCmd)
}

func printSSEEvent(eventType string, rawData string) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(rawData), &data); err != nil {
		fmt.Printf("[%s] %s\n", eventType, rawData)
		return
	}

	parts := []string{"[" + eventType + "]"}
	if runID, ok := data["run_id"].(string); ok && runID != "" {
		parts = append(parts, "run="+runID)
	}
	if job, ok := data["job"].(string); ok && job != "" {
		parts = append(parts, "job="+job)
	}
	if level, ok := data["level"].(string); ok && level != "" {
		parts = append(parts, "level="+level)
	}

	message := rawData
	if value, ok := data["message"].(string); ok && value != "" {
		message = value
	}

	fmt.Printf("%s %s\n", strings.Join(parts, " "), message)
}

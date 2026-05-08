package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/muhammedikinci/pin/internal/runner"
	"github.com/spf13/cobra"
)

var doctorFilePath string
var doctorURL string
var doctorToken string
var doctorPort int

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local Docker, config, and daemon access",
	Long:  "Run quick diagnostics for local pipeline execution and VPS daemon connectivity.",
	RunE: func(cmd *cobra.Command, args []string) error {
		failures := 0

		if err := checkDocker(); err != nil {
			failures++
			printCheck("Docker daemon", err)
		} else {
			printCheck("Docker daemon", nil)
		}

		if doctorFilePath != "" {
			if err := runner.Validate(doctorFilePath); err != nil {
				failures++
				printCheck("Pipeline config", err)
			} else {
				printCheck("Pipeline config", nil)
			}
		}

		if doctorPort > 0 {
			if err := checkPort(doctorPort); err != nil {
				failures++
				printCheck(fmt.Sprintf("Local port %d", doctorPort), err)
			} else {
				printCheck(fmt.Sprintf("Local port %d", doctorPort), nil)
			}
		}

		if doctorURL != "" {
			if err := checkDaemonHealth(doctorURL, doctorToken); err != nil {
				failures++
				printCheck("Remote daemon", err)
			} else {
				printCheck("Remote daemon", nil)
			}
		}

		if failures > 0 {
			return fmt.Errorf("doctor found %d issue(s)", failures)
		}
		return nil
	},
}

func init() {
	doctorCmd.Flags().StringVarP(&doctorFilePath, "filepath", "f", "", "pipeline configuration file path to validate")
	doctorCmd.Flags().StringVar(&doctorURL, "url", "", "Pin daemon URL to check")
	doctorCmd.Flags().StringVar(&doctorToken, "token", os.Getenv("PIN_TOKEN"), "bearer token; defaults to PIN_TOKEN")
	doctorCmd.Flags().IntVar(&doctorPort, "port", 0, "optional local daemon port to check")
	rootCmd.AddCommand(doctorCmd)
}

func checkDocker() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	defer cli.Close()

	_, err = cli.Ping(ctx)
	return err
}

func checkPort(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	return listener.Close()
}

func checkDaemonHealth(rawURL string, token string) error {
	endpoint, err := normalizeRemoteEndpoint(rawURL, "/health")
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	addToken(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("daemon returned %s", resp.Status)
	}
	return nil
}

func printCheck(name string, err error) {
	if err != nil {
		fmt.Printf("[fail] %s: %v\n", name, err)
		return
	}
	fmt.Printf("[ok] %s\n", name)
}

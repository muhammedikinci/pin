package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	pathfile "path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	pinerrors "github.com/muhammedikinci/pin/internal/errors"
	"github.com/muhammedikinci/pin/internal/sse"
	"github.com/spf13/viper"
)

var configMutex sync.Mutex

type DaemonOptions struct {
	Host               string
	Port               int
	Token              string
	MaxConcurrent      int
	DataDir            string
	GitHubSecret       string
	GitHubBranch       string
	GitHubPipelineFile string
}

func Apply(filepath string) error {
	pipeline, err := loadPipelineFromFile(filepath)
	if err != nil {
		printPipelineSetupError(err)
		return err
	}

	color.Set(color.FgGreen)
	fmt.Println("✅ Pipeline validation successful")
	color.Unset()

	currentRunner := Runner{}

	if err := currentRunner.run(pipeline); err != nil {
		// Enhanced error handling for execution errors
		if pinErr, ok := err.(*pinerrors.PinError); ok {
			fmt.Print(pinerrors.ConsoleFormatter.Format(pinErr))
		} else {
			// Create enhanced error for unknown execution errors
			execErr := pinerrors.NewPinError(pinerrors.ErrCodeJobExecution, "pipeline execution failed").
				WithCause(err).
				AddSuggestions(
					"Check Docker daemon is running",
					"Verify all required images are available",
					"Review script commands for errors",
					"Enable verbose logging with 'logsWithTime: true'",
				)
			fmt.Print(pinerrors.ConsoleFormatter.Format(execErr))
		}
		return err
	}

	color.Unset()
	return nil
}

func Validate(filepath string) error {
	_, err := loadPipelineFromFile(filepath)
	return err
}

func loadPipelineFromFile(filepath string) (Pipeline, error) {
	if err := checkFileExists(filepath); err != nil {
		return Pipeline{}, err
	}

	configMutex.Lock()
	defer configMutex.Unlock()

	if err := readConfig(filepath); err != nil {
		return Pipeline{}, err
	}

	return validateAndParseCurrentConfig()
}

func loadPipelineFromYAML(yamlContent []byte) (Pipeline, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	viper.Reset()
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(yamlContent)); err != nil {
		return Pipeline{}, fmt.Errorf("failed to parse YAML configuration: %w", err)
	}

	return validateAndParseCurrentConfig()
}

func validateAndParseCurrentConfig() (Pipeline, error) {
	validator := NewPipelineValidator()
	if err := validator.ValidatePipeline(); err != nil {
		return Pipeline{}, err
	}

	pipeline, err := parse()
	if err != nil {
		if pinErr, ok := err.(*pinerrors.PinError); ok {
			return Pipeline{}, pinErr
		}
		return Pipeline{}, pinerrors.NewPinError(pinerrors.ErrCodePipelineValidation, "failed to parse pipeline configuration").
			WithCause(err).
			AddSuggestions(
				"Check YAML syntax and formatting",
				"Ensure all required fields are present",
				"Validate YAML using an online validator",
			)
	}

	return pipeline, nil
}

func printPipelineSetupError(err error) {
	if pinErr, ok := err.(*pinerrors.PinError); ok {
		fmt.Print(pinerrors.ConsoleFormatter.Format(pinErr))
		return
	}

	color.Set(color.FgRed)
	fmt.Printf("Pipeline validation failed: %s\n", err.Error())
	color.Unset()
}

func checkFileExists(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		fileBuilder := pinerrors.NewFileErrorBuilder()
		return fileBuilder.FileNotFound(filepath, err)
	}

	return nil
}

func readConfig(filepath string) error {
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsPermission(err) {
			fileBuilder := pinerrors.NewFileErrorBuilder()
			return fileBuilder.PermissionDenied(filepath, err)
		}
		fileBuilder := pinerrors.NewFileErrorBuilder()
		return fileBuilder.FileNotFound(filepath, err)
	}

	viper.Reset()
	viper.SetConfigType("yaml")

	err = viper.ReadConfig(bytes.NewBuffer(fileBytes))
	if err != nil {
		return pinerrors.NewPinError(pinerrors.ErrCodeInvalidConfig, "failed to parse YAML configuration").
			WithFile(filepath).
			WithCause(err).
			AddSuggestions(
				"Check YAML syntax - ensure proper indentation",
				"Validate YAML format using an online validator",
				"Ensure no tabs are used (use spaces for indentation)",
				"Check for missing quotes around strings with special characters",
			)
	}

	return nil
}

// executeYAMLPipeline executes a pipeline from YAML content
func executeYAMLPipeline(yamlContent []byte, runID string, logWriter io.Writer) error {
	pipeline, err := loadPipelineFromYAML(yamlContent)
	if err != nil {
		return err
	}
	pipeline.RunID = runID
	pipeline.LogWriter = logWriter

	currentRunner := Runner{}
	if err := currentRunner.run(pipeline); err != nil {
		return fmt.Errorf("pipeline execution failed: %w", err)
	}

	return nil
}

// ApplyDaemon runs the application in daemon mode with SSE server
func ApplyDaemon(filepath string) error {
	return ApplyDaemonWithOptions(filepath, DaemonOptions{
		Host:          "127.0.0.1",
		Port:          8081,
		MaxConcurrent: 1,
	})
}

func ApplyDaemonWithOptions(filepath string, options DaemonOptions) error {
	if options.Host == "" {
		options.Host = "127.0.0.1"
	}
	if options.Port == 0 {
		options.Port = 8081
	}
	if options.MaxConcurrent < 1 {
		options.MaxConcurrent = 1
	}
	if options.DataDir == "" {
		options.DataDir = ".pin"
	}

	log.Printf("Starting PIN in daemon mode on %s:%d...", options.Host, options.Port)
	if options.Token == "" && (options.Host == "0.0.0.0" || options.Host == "::") {
		log.Printf("WARNING: daemon is exposed without a token. Set PIN_TOKEN or pass --token before opening it to the internet.")
	}

	// Create event broadcaster
	broadcaster := sse.NewEventBroadcaster()
	sse.SetGlobalBroadcaster(broadcaster)
	defer sse.SetGlobalBroadcaster(nil)

	// Set pipeline executor function to handle HTTP triggered pipelines
	sse.SetPipelineExecutor(func(yamlContent []byte, execution sse.PipelineExecution) error {
		return executeYAMLPipeline(yamlContent, execution.RunID, execution.LogWriter)
	})
	defer sse.SetPipelineExecutor(nil)

	// Create and start SSE server
	sseServer := sse.NewServerWithOptions(sse.ServerOptions{
		Host:            options.Host,
		Port:            options.Port,
		Token:           options.Token,
		MaxConcurrent:   options.MaxConcurrent,
		DataDir:         pathfile.Join(options.DataDir, "runs"),
		GitHubSecret:    options.GitHubSecret,
		GitHubBranch:    options.GitHubBranch,
		WebhookPipeline: options.GitHubPipelineFile,
	}, broadcaster, log.New(os.Stdout, "[SSE] ", log.LstdFlags))

	// Note: Context for graceful shutdown is handled by signal handling

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start SSE server in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("SSE server starting on %s:%d", options.Host, options.Port)
		if err := sseServer.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Printf("SSE server error: %v", err)
		}
	}()

	// Broadcast daemon start event
	broadcaster.Broadcast(sse.Event{
		Type: "daemon_start",
		Data: map[string]interface{}{
			"message":         "PIN daemon started successfully",
			"sse_endpoint":    fmt.Sprintf("http://%s:%d/events", displayHost(options.Host), options.Port),
			"health_endpoint": fmt.Sprintf("http://%s:%d/health", displayHost(options.Host), options.Port),
			"runs_endpoint":   fmt.Sprintf("http://%s:%d/runs", displayHost(options.Host), options.Port),
		},
		Timestamp: time.Now(),
	})

	// If a filepath was provided, run the pipeline immediately
	if filepath != "" {
		log.Printf("Running initial pipeline from: %s", filepath)
		go func() {
			if err := Apply(filepath); err != nil {
				log.Printf("Initial pipeline failed: %v", err)
				broadcaster.Broadcast(sse.Event{
					Type: "pipeline_error",
					Data: map[string]interface{}{
						"message": "Initial pipeline execution failed",
						"error":   err.Error(),
						"file":    filepath,
					},
					Timestamp: time.Now(),
				})
			} else {
				broadcaster.Broadcast(sse.Event{
					Type: "pipeline_complete",
					Data: map[string]interface{}{
						"message": "Initial pipeline execution completed successfully",
						"file":    filepath,
					},
					Timestamp: time.Now(),
				})
			}
		}()
	}

	// Wait for shutdown signal
	<-sigChan
	log.Printf("Received shutdown signal, gracefully shutting down...")

	// Broadcast daemon stop event
	broadcaster.Broadcast(sse.Event{
		Type: "daemon_stop",
		Data: map[string]interface{}{
			"message": "PIN daemon shutting down",
		},
		Timestamp: time.Now(),
	})

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop SSE server
	if err := sseServer.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping SSE server: %v", err)
	}

	// Close broadcaster
	broadcaster.Close()

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("PIN daemon stopped gracefully")
	case <-shutdownCtx.Done():
		log.Printf("PIN daemon shutdown timeout")
	}

	return nil
}

func displayHost(host string) string {
	if host == "" || host == "0.0.0.0" || host == "::" {
		return "localhost"
	}
	return host
}

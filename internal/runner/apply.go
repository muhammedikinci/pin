package runner

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	pinerrors "github.com/muhammedikinci/pin/internal/errors"
	"github.com/muhammedikinci/pin/internal/interfaces"
	"github.com/muhammedikinci/pin/internal/sse"
	"github.com/spf13/viper"
)

func Apply(filepath string) error {
	if err := checkFileExists(filepath); err != nil {
		return err
	}

	if err := readConfig(filepath); err != nil {
		return err
	}

	// Validate pipeline configuration before execution
	validator := NewPipelineValidator()
	if err := validator.ValidatePipeline(); err != nil {
		// Format and display the error using our enhanced error reporting
		if pinErr, ok := err.(*pinerrors.PinError); ok {
			fmt.Print(pinerrors.ConsoleFormatter.Format(pinErr))
		} else {
			// Fallback for non-PinError errors
			color.Set(color.FgRed)
			fmt.Printf("Pipeline validation failed: %s\n", err.Error())
			color.Unset()
		}
		return err
	}

	color.Set(color.FgGreen)
	fmt.Println("✅ Pipeline validation successful")
	color.Unset()

	pipeline, err := parse()
	if err != nil {
		// Enhanced error handling for parse errors
		if pinErr, ok := err.(*pinerrors.PinError); ok {
			fmt.Print(pinerrors.ConsoleFormatter.Format(pinErr))
		} else {
			// Create enhanced error for unknown parse errors
			parseErr := pinerrors.NewPinError(pinerrors.ErrCodePipelineValidation, "failed to parse pipeline configuration").
				WithCause(err).
				AddSuggestions(
					"Check YAML syntax and formatting",
					"Ensure all required fields are present",
					"Validate YAML using an online validator",
				)
			fmt.Print(pinerrors.ConsoleFormatter.Format(parseErr))
		}
		return err
	}

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
func executeYAMLPipeline(yamlContent []byte) error {
	// Configure viper to read YAML from the provided content
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer(yamlContent))
	if err != nil {
		return fmt.Errorf("failed to parse YAML configuration: %w", err)
	}

	// Validate pipeline configuration before execution
	validator := NewPipelineValidator()
	if err := validator.ValidatePipeline(); err != nil {
		return fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Parse and run the pipeline
	pipeline, err := parse()
	if err != nil {
		return fmt.Errorf("failed to parse pipeline: %w", err)
	}

	currentRunner := Runner{}
	if err := currentRunner.run(pipeline); err != nil {
		return fmt.Errorf("pipeline execution failed: %w", err)
	}

	return nil
}

// ApplyDaemon runs the application in daemon mode with SSE server
func ApplyDaemon(filepath string) error {
	log.Printf("Starting PIN in daemon mode...")

	// Create event broadcaster
	broadcaster := sse.NewEventBroadcaster()
	sse.SetGlobalBroadcaster(broadcaster)

	// Set pipeline executor function to handle HTTP triggered pipelines
	sse.SetPipelineExecutor(func(yamlContent []byte) error {
		return executeYAMLPipeline(yamlContent)
	})

	// Create and start SSE server
	sseServer := sse.NewServer(8081, broadcaster, log.New(os.Stdout, "[SSE] ", log.LstdFlags))

	// Note: Context for graceful shutdown is handled by signal handling

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start SSE server in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("SSE server starting on :8081")
		if err := sseServer.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Printf("SSE server error: %v", err)
		}
	}()

	// Broadcast daemon start event
	broadcaster.Broadcast(interfaces.Event{
		Type: "daemon_start",
		Data: map[string]interface{}{
			"message":         "PIN daemon started successfully",
			"sse_endpoint":    "http://localhost:8081/events",
			"health_endpoint": "http://localhost:8081/health",
		},
		Timestamp: time.Now(),
	})

	// If a filepath was provided, run the pipeline immediately
	if filepath != "" {
		log.Printf("Running initial pipeline from: %s", filepath)
		go func() {
			if err := Apply(filepath); err != nil {
				log.Printf("Initial pipeline failed: %v", err)
				broadcaster.Broadcast(interfaces.Event{
					Type: "pipeline_error",
					Data: map[string]interface{}{
						"message": "Initial pipeline execution failed",
						"error":   err.Error(),
						"file":    filepath,
					},
					Timestamp: time.Now(),
				})
			} else {
				broadcaster.Broadcast(interfaces.Event{
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
	broadcaster.Broadcast(interfaces.Event{
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

package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// PipelineExecutor is a function type for executing pipelines from YAML
type PipelineExecutor func(yamlContent []byte) error

// Global pipeline executor function
var pipelineExecutor PipelineExecutor

// SetPipelineExecutor sets the global pipeline executor function
func SetPipelineExecutor(executor PipelineExecutor) {
	pipelineExecutor = executor
}

// Server represents an SSE server that can broadcast events to connected clients
type Server struct {
	broadcaster EventBroadcaster
	server      *http.Server
	logger      *log.Logger
}

// NewServer creates a new SSE server instance
func NewServer(port int, broadcaster EventBroadcaster, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.New(log.Writer(), "[SSE] ", log.LstdFlags)
	}

	server := &Server{
		broadcaster: broadcaster,
		logger:      logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.handleSSE)
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/trigger", server.handleTrigger)
	mux.HandleFunc("/", server.handleRoot)

	server.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.corsMiddleware(mux),
	}

	return server
}

// Start starts the SSE server
func (s *Server) Start() error {
	s.logger.Printf("Starting SSE server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop stops the SSE server gracefully
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Println("Stopping SSE server...")
	return s.server.Shutdown(ctx)
}

// handleSSE handles Server-Sent Events connections
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create client channel
	clientChan := make(chan Event, 100) // Buffer for 100 events
	clientID := s.broadcaster.AddClient(clientChan)

	if clientID == "" {
		http.Error(w, "Failed to register SSE client", http.StatusInternalServerError)
		return
	}

	s.logger.Printf("New SSE client connected: %s", clientID)

	// Clean up when client disconnects
	defer func() {
		s.broadcaster.RemoveClient(clientID)
		s.logger.Printf("SSE client disconnected: %s", clientID)
	}()

	// Create a flusher to send events immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Listen for events and client disconnect
	for {
		select {
		case event, ok := <-clientChan:
			if !ok {
				// Channel closed, client should disconnect
				return
			}

			// Convert event to JSON
			eventData, err := json.Marshal(event.Data)
			if err != nil {
				s.logger.Printf("Error marshaling event data: %v", err)
				continue
			}

			// Send SSE formatted event
			fmt.Fprintf(w, "id: %s\n", event.ID)
			fmt.Fprintf(w, "event: %s\n", event.Type)
			fmt.Fprintf(w, "data: %s\n\n", string(eventData))
			flusher.Flush()

		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// handleHealth provides a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"status":     "healthy",
		"clients":    s.broadcaster.GetClientCount(),
		"timestamp":  time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// handleRoot provides information about available endpoints
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"service":   "PIN SSE Server",
		"version":   "1.0.0",
		"endpoints": map[string]string{
			"/events":  "Server-Sent Events endpoint for real-time pipeline updates",
			"/health":  "Health check endpoint",
			"/trigger": "POST endpoint to trigger pipeline execution with YAML configuration",
		},
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// handleTrigger handles POST requests to trigger pipeline execution
func (s *Server) handleTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the YAML configuration from request body
	yamlContent, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(yamlContent) == 0 {
		http.Error(w, "Empty YAML configuration", http.StatusBadRequest)
		return
	}

	s.logger.Printf("Received pipeline trigger request")

	// Broadcast pipeline trigger event
	if s.broadcaster != nil {
		s.broadcaster.Broadcast(Event{
			Type: "pipeline_trigger",
			Data: map[string]interface{}{
				"message": "Pipeline trigger request received",
				"source":  "http_endpoint",
			},
			Timestamp: time.Now(),
		})
	}

	// Execute pipeline in goroutine to avoid blocking the HTTP request
	go func() {
		if err := s.executePipelineFromYAML(yamlContent); err != nil {
			s.logger.Printf("Pipeline execution failed: %v", err)
			if s.broadcaster != nil {
				s.broadcaster.Broadcast(Event{
					Type: "pipeline_error",
					Data: map[string]interface{}{
						"message": "Pipeline execution failed",
						"error":   err.Error(),
						"source":  "http_endpoint",
					},
					Timestamp: time.Now(),
				})
			}
		} else {
			s.logger.Printf("Pipeline execution completed successfully")
			if s.broadcaster != nil {
				s.broadcaster.Broadcast(Event{
					Type: "pipeline_complete",
					Data: map[string]interface{}{
						"message": "Pipeline execution completed successfully",
						"source":  "http_endpoint",
					},
					Timestamp: time.Now(),
				})
			}
		}
	}()

	// Return immediate response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "accepted",
		"message":   "Pipeline execution started",
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// executePipelineFromYAML executes a pipeline from YAML configuration
func (s *Server) executePipelineFromYAML(yamlContent []byte) error {
	// We need to move this functionality to avoid import cycle
	// For now, we'll store the YAML and trigger execution via the apply package
	// The actual execution will be handled by the runner package through a callback mechanism

	if pipelineExecutor != nil {
		return pipelineExecutor(yamlContent)
	}

	return fmt.Errorf("pipeline executor not configured")
}

// corsMiddleware adds CORS headers to allow web clients to connect
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
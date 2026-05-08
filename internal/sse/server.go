package sse

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// PipelineExecutor is a function type for executing pipelines from YAML
type PipelineExecutor func(yamlContent []byte, execution PipelineExecution) error

type PipelineExecution struct {
	RunID     string
	LogWriter io.Writer
}

// Global pipeline executor function
var pipelineExecutor PipelineExecutor

// SetPipelineExecutor sets the global pipeline executor function
func SetPipelineExecutor(executor PipelineExecutor) {
	pipelineExecutor = executor
}

// Server represents an SSE server that can broadcast events to connected clients
type Server struct {
	broadcaster       EventBroadcaster
	server            *http.Server
	logger            *log.Logger
	token             string
	runStore          *RunStore
	concurrentLimiter chan struct{}
	githubSecret      string
	githubBranch      string
	webhookPipeline   string
}

type ServerOptions struct {
	Host          string
	Port          int
	Token         string
	MaxConcurrent int
	DataDir       string

	GitHubSecret    string
	GitHubBranch    string
	WebhookPipeline string
}

// NewServer creates a new SSE server instance
func NewServer(port int, broadcaster EventBroadcaster, logger *log.Logger) *Server {
	return NewServerWithOptions(ServerOptions{Port: port, MaxConcurrent: 1}, broadcaster, logger)
}

func NewServerWithOptions(options ServerOptions, broadcaster EventBroadcaster, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.New(log.Writer(), "[SSE] ", log.LstdFlags)
	}
	if options.Port == 0 {
		options.Port = 8081
	}
	if options.MaxConcurrent < 1 {
		options.MaxConcurrent = 1
	}

	server := &Server{
		broadcaster:       broadcaster,
		logger:            logger,
		token:             options.Token,
		runStore:          NewRunStoreWithDir(50, options.DataDir),
		concurrentLimiter: make(chan struct{}, options.MaxConcurrent),
		githubSecret:      options.GitHubSecret,
		githubBranch:      options.GitHubBranch,
		webhookPipeline:   options.WebhookPipeline,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.handleSSE)
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/trigger", server.handleTrigger)
	mux.HandleFunc("/runs", server.handleRuns)
	mux.HandleFunc("/runs/", server.handleRunDetail)
	mux.HandleFunc("/webhooks/github", server.handleGitHubWebhook)
	mux.HandleFunc("/", server.handleRoot)

	server.server = &http.Server{
		Addr:    net.JoinHostPort(options.Host, strconv.Itoa(options.Port)),
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
	if !s.authorize(w, r) {
		return
	}

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
	if !s.authorize(w, r) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":    "healthy",
		"clients":   s.broadcaster.GetClientCount(),
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// handleRoot provides information about available endpoints
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"service": "PIN SSE Server",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"/events":          "Server-Sent Events endpoint for real-time pipeline updates",
			"/health":          "Health check endpoint",
			"/trigger":         "POST endpoint to trigger pipeline execution with YAML configuration",
			"/runs":            "Recent pipeline run statuses",
			"/webhooks/github": "GitHub push webhook endpoint",
		},
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// handleTrigger handles POST requests to trigger pipeline execution
func (s *Server) handleTrigger(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r) {
		return
	}

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
	run := s.queuePipeline(yamlContent, "http_endpoint")

	// Return immediate response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "accepted",
		"message":   "Pipeline execution queued",
		"run_id":    run.ID,
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"runs":      s.runStore.List(),
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleRunDetail(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/runs/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	if strings.HasSuffix(path, "/logs") {
		runID := strings.TrimSuffix(path, "/logs")
		s.handleRunLogs(w, r, runID)
		return
	}

	record, ok := s.runStore.Get(path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func (s *Server) handleRunLogs(w http.ResponseWriter, r *http.Request, runID string) {
	if _, ok := s.runStore.Get(runID); !ok {
		http.NotFound(w, r)
		return
	}

	logContent, err := s.runStore.ReadLog(runID)
	if err != nil {
		http.Error(w, "run log not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(logContent)
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.githubSecret == "" {
		http.Error(w, "GitHub webhook secret is not configured", http.StatusServiceUnavailable)
		return
	}
	if s.webhookPipeline == "" {
		http.Error(w, "GitHub webhook pipeline is not configured", http.StatusServiceUnavailable)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !verifyGitHubSignature(payload, r.Header.Get("X-Hub-Signature-256"), s.githubSecret) {
		http.Error(w, "Invalid GitHub signature", http.StatusUnauthorized)
		return
	}

	var event struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		http.Error(w, "Invalid GitHub webhook payload", http.StatusBadRequest)
		return
	}

	branch := branchFromRef(event.Ref)
	if s.githubBranch != "" && branch != s.githubBranch {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "skipped",
			"message":   "branch filter did not match",
			"branch":    branch,
			"timestamp": time.Now(),
		})
		return
	}

	yamlContent, err := os.ReadFile(s.webhookPipeline)
	if err != nil {
		http.Error(w, "Failed to read webhook pipeline file", http.StatusInternalServerError)
		return
	}

	run := s.queuePipeline(yamlContent, "github_webhook")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "accepted",
		"message":   "GitHub webhook pipeline queued",
		"branch":    branch,
		"run_id":    run.ID,
		"timestamp": time.Now(),
	})
}

func (s *Server) queuePipeline(yamlContent []byte, source string) RunRecord {
	run := s.runStore.StartRun(source)

	s.broadcast(Event{
		Type: "pipeline_trigger",
		Data: map[string]interface{}{
			"message": "Pipeline trigger request received",
			"source":  source,
			"run_id":  run.ID,
		},
		Timestamp: time.Now(),
	})

	go s.executeQueuedPipeline(run, yamlContent, source)

	return run
}

func (s *Server) executeQueuedPipeline(run RunRecord, yamlContent []byte, source string) {
	s.concurrentLimiter <- struct{}{}
	defer func() {
		<-s.concurrentLimiter
	}()

	s.runStore.MarkRunning(run.ID)
	s.broadcast(Event{
		Type: "pipeline_started",
		Data: map[string]interface{}{
			"message": "Pipeline execution started",
			"source":  source,
			"run_id":  run.ID,
		},
		Timestamp: time.Now(),
	})

	execution := PipelineExecution{
		RunID:     run.ID,
		LogWriter: s.runStore.NewLogWriter(run.ID),
	}

	if err := s.executePipelineFromYAML(yamlContent, execution); err != nil {
		s.runStore.MarkFailed(run.ID, err)
		s.logger.Printf("Pipeline execution failed: %v", err)
		s.broadcast(Event{
			Type: "pipeline_error",
			Data: map[string]interface{}{
				"message": "Pipeline execution failed",
				"error":   err.Error(),
				"source":  source,
				"run_id":  run.ID,
			},
			Timestamp: time.Now(),
		})
		return
	}

	s.runStore.MarkSucceeded(run.ID)
	s.logger.Printf("Pipeline execution completed successfully")
	s.broadcast(Event{
		Type: "pipeline_complete",
		Data: map[string]interface{}{
			"message": "Pipeline execution completed successfully",
			"source":  source,
			"run_id":  run.ID,
		},
		Timestamp: time.Now(),
	})
}

func (s *Server) broadcast(event Event) {
	s.runStore.AppendEvent(event)
	if s.broadcaster != nil {
		s.broadcaster.Broadcast(event)
	}
}

// executePipelineFromYAML executes a pipeline from YAML configuration
func (s *Server) executePipelineFromYAML(yamlContent []byte, execution PipelineExecution) error {
	// We need to move this functionality to avoid import cycle
	// For now, we'll store the YAML and trigger execution via the apply package
	// The actual execution will be handled by the runner package through a callback mechanism

	if pipelineExecutor != nil {
		return pipelineExecutor(yamlContent, execution)
	}

	return fmt.Errorf("pipeline executor not configured")
}

func verifyGitHubSignature(payload []byte, signatureHeader string, secret string) bool {
	if signatureHeader == "" || secret == "" {
		return false
	}
	signatureHeader = strings.TrimSpace(signatureHeader)
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}

	expectedMAC := hmac.New(sha256.New, []byte(secret))
	expectedMAC.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(expectedMAC.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(signatureHeader))
}

func branchFromRef(ref string) string {
	const prefix = "refs/heads/"
	if strings.HasPrefix(ref, prefix) {
		return strings.TrimPrefix(ref, prefix)
	}
	return ref
}

// corsMiddleware adds CORS headers to allow web clients to connect
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Pin-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) authorize(w http.ResponseWriter, r *http.Request) bool {
	if s.token == "" {
		return true
	}

	if r.URL.Query().Get("token") == s.token {
		return true
	}

	if r.Header.Get("X-Pin-Token") == s.token {
		return true
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token := strings.TrimSpace(authHeader[len("Bearer "):])
		if token == s.token {
			return true
		}
	}

	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	return false
}

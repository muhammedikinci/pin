package sse

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServer_HandleHealth(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServer(8081, mockBroadcaster, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestServer_HandleRoot(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServer(8081, mockBroadcaster, nil)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestServer_HandleTrigger_InvalidMethod(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServer(8081, mockBroadcaster, nil)

	req := httptest.NewRequest("GET", "/trigger", nil)
	w := httptest.NewRecorder()

	server.handleTrigger(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestServer_HandleTrigger_EmptyBody(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServer(8081, mockBroadcaster, nil)

	req := httptest.NewRequest("POST", "/trigger", bytes.NewReader([]byte("")))
	w := httptest.NewRecorder()

	server.handleTrigger(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestServer_HandleTrigger_ValidYAML(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServer(8081, mockBroadcaster, nil)

	// Set a test pipeline executor that always succeeds
	SetPipelineExecutor(func(yamlContent []byte, execution PipelineExecution) error {
		return nil
	})

	yamlContent := `
workflow:
  - test_job

test_job:
  image: "alpine:latest"
  script:
    - "echo test"
`

	req := httptest.NewRequest("POST", "/trigger", bytes.NewReader([]byte(yamlContent)))
	w := httptest.NewRecorder()

	server.handleTrigger(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Give some time for the goroutine to complete
	time.Sleep(100 * time.Millisecond)
}

func TestServer_HandleRuns(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServer(8081, mockBroadcaster, nil)

	run := server.runStore.StartRun("test")
	server.runStore.MarkRunning(run.ID)

	req := httptest.NewRequest("GET", "/runs", nil)
	w := httptest.NewRecorder()

	server.handleRuns(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !bytes.Contains(w.Body.Bytes(), []byte(run.ID)) {
		t.Errorf("Expected response to contain run ID %s, got %s", run.ID, w.Body.String())
	}
}

func TestServer_AuthRequired(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServerWithOptions(ServerOptions{
		Port:  8081,
		Token: "secret",
	}, mockBroadcaster, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestServer_AuthAllowsBearerToken(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServerWithOptions(ServerOptions{
		Port:  8081,
		Token: "secret",
	}, mockBroadcaster, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestServer_HandleGitHubWebhook(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	pipelinePath := filepath.Join(t.TempDir(), "pin.yaml")
	pipeline := []byte("workflow:\n  - test\n\ntest:\n  image: alpine:latest\n")
	if err := os.WriteFile(pipelinePath, pipeline, 0644); err != nil {
		t.Fatalf("failed to write pipeline file: %v", err)
	}

	executed := make(chan PipelineExecution, 1)
	SetPipelineExecutor(func(yamlContent []byte, execution PipelineExecution) error {
		executed <- execution
		return nil
	})

	server := NewServerWithOptions(ServerOptions{
		Port:            8081,
		DataDir:         t.TempDir(),
		GitHubSecret:    "secret",
		GitHubBranch:    "main",
		WebhookPipeline: pipelinePath,
	}, mockBroadcaster, nil)

	payload := []byte(`{"ref":"refs/heads/main"}`)
	req := httptest.NewRequest("POST", "/webhooks/github", bytes.NewReader(payload))
	req.Header.Set("X-Hub-Signature-256", githubSignature(payload, "secret"))
	w := httptest.NewRecorder()

	server.handleGitHubWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	select {
	case execution := <-executed:
		if execution.RunID == "" {
			t.Fatal("expected webhook execution to include run id")
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected webhook pipeline to execute")
	}
}

func TestServer_CorsMiddleware(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServer(8081, mockBroadcaster, nil)

	handler := server.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check CORS headers
	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin: *, got %s", origin)
	}

	if methods := w.Header().Get("Access-Control-Allow-Methods"); methods != "GET, POST, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods: GET, POST, OPTIONS, got %s", methods)
	}
}

func githubSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestServer_CorsMiddleware_Options(t *testing.T) {
	mockBroadcaster := NewEventBroadcaster()
	server := NewServer(8081, mockBroadcaster, nil)

	handler := server.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for OPTIONS request")
	}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS request, got %d", w.Code)
	}
}

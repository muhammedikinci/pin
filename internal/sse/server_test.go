package sse

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/muhammedikinci/pin/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestServer_HandleHealth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)
	server := NewServer(8081, mockBroadcaster, nil)

	req := httptest.NewRequest("GET", "/trigger", nil)
	w := httptest.NewRecorder()

	server.handleTrigger(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestServer_HandleTrigger_EmptyBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)
	server := NewServer(8081, mockBroadcaster, nil)

	req := httptest.NewRequest("POST", "/trigger", bytes.NewReader([]byte("")))
	w := httptest.NewRecorder()

	server.handleTrigger(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestServer_HandleTrigger_ValidYAML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)

	// Expect broadcast calls for pipeline trigger event
	mockBroadcaster.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(event interface{}) {
			// We expect at least one broadcast call for the trigger event
		}).
		MinTimes(1)

	server := NewServer(8081, mockBroadcaster, nil)

	// Set a test pipeline executor that always succeeds
	SetPipelineExecutor(func(yamlContent []byte) error {
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

func TestServer_CorsMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)
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

func TestServer_CorsMiddleware_Options(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcaster := mocks.NewMockEventBroadcaster(ctrl)
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
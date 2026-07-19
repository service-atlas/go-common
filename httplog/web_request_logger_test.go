package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// captureState holds shared state for the capturingHandler so children produced by
// WithAttrs write into the same storage.
type captureState struct {
	lastAttrs []slog.Attr
}

type capturingHandler struct {
	state *captureState
	attrs []slog.Attr
}

func newCapturingHandler() *capturingHandler { return &capturingHandler{state: &captureState{}} }

func (h *capturingHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }

func (h *capturingHandler) Handle(ctx context.Context, r slog.Record) error {
	merged := make([]slog.Attr, 0, len(h.attrs)+4)
	merged = append(merged, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		merged = append(merged, a)
		return true
	})
	h.state.lastAttrs = merged
	return nil
}

func (h *capturingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Create a child that shares the same state but has accumulated attrs.
	acc := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	acc = append(acc, h.attrs...)
	acc = append(acc, attrs...)
	return &capturingHandler{state: h.state, attrs: acc}
}

func (h *capturingHandler) WithGroup(name string) slog.Handler { return h }

func (h *capturingHandler) lastAllAttrs() []slog.Attr {
	return append([]slog.Attr{}, h.state.lastAttrs...)
}

func TestLoggerFromContext_Default(t *testing.T) {
	ctx := context.Background()
	l := LoggerFromContext(ctx)
	if l != slog.Default() {
		t.Fatalf("expected default logger when none set in context")
	}
}

func TestWebRequestLogger_RequestID(t *testing.T) {
	cap := newCapturingHandler()
	orig := slog.Default()
	slog.SetDefault(slog.New(cap))
	t.Cleanup(func() { slog.SetDefault(orig) })

	expectedID := "req-12345"
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Request id should be available via helper
		if got := GetRequestId(r); got != expectedID {
			t.Errorf("expected GetRequestId to return %q, got %q", expectedID, got)
		}
		// Logger with request_id attribute should be in context
		LoggerFromContext(r.Context()).Info("test log")
		w.WriteHeader(http.StatusOK)
	})

	h := WebRequestLogger(next)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Request-Id", expectedID)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, r)

	attrs := cap.lastAllAttrs()
	found := false
	for _, a := range attrs {
		if a.Key == requestKey && a.Value.String() == expectedID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("logger did not include attr %q=%q; got attrs: %#v", requestKey, expectedID, attrs)
	}
}

func TestWebRequestLogger_GeneratesUUID(t *testing.T) {
	cap := newCapturingHandler()
	orig := slog.Default()
	slog.SetDefault(slog.New(cap))
	t.Cleanup(func() { slog.SetDefault(orig) })

	var capturedID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestId(r)
		LoggerFromContext(r.Context()).Info("test log")
		w.WriteHeader(http.StatusOK)
	})

	h := WebRequestLogger(next)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, r)

	if _, err := uuid.Parse(capturedID); err != nil {
		t.Fatalf("expected generated request id to be a UUID, got %q: %v", capturedID, err)
	}

	attrs := cap.lastAllAttrs()
	found := false
	for _, a := range attrs {
		if a.Key == requestKey && a.Value.String() == capturedID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("logger did not include generated request id attr %q=%q; got attrs: %#v", requestKey, capturedID, attrs)
	}
}

func TestGetRequestId_EmptyWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := GetRequestId(r); got != "" {
		t.Fatalf("expected empty string when request id not in context, got %q", got)
	}
}

func TestStructuredLogger(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a JSON handler that writes to our buffer
	jsonHandler := slog.NewJSONHandler(&logBuffer, nil)
	oldDefault := slog.Default()
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(oldDefault) })
	// Create a test HTTP handler that returns a specific status code
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("test response"))
	})

	// Wrap our test handler with the StructuredLogger middleware
	middlewareHandler := WebRequestLogger(testHandler)

	// Create a test request
	req := httptest.NewRequest("POST", "/test-path", nil)
	req.Header.Set("User-Agent", "test-agent")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	middlewareHandler.ServeHTTP(rr, req)

	// Check the response
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if body := rr.Body.String(); body != "test response" {
		t.Errorf("handler returned unexpected body: got %v want %v", body, "test response")
	}

	// Parse the log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(logBuffer.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify log fields
	if msg, ok := logEntry["msg"].(string); !ok || msg != "WEB_REQUEST" {
		t.Errorf("Expected log message 'WEB_REQUEST', got %v", msg)
	}

	if id, ok := logEntry[requestKey].(string); !ok || id == "" {
		t.Errorf("Expected non-empty %q in log, got %v", requestKey, id)
	}

	if method, ok := logEntry["method"].(string); !ok || method != "POST" {
		t.Errorf("Expected method 'POST', got %v", method)
	}

	if path, ok := logEntry["path"].(string); !ok || path != "/test-path" {
		t.Errorf("Expected path '/test-path', got %v", path)
	}

	if status, ok := logEntry["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Errorf("Expected status %d, got %v", http.StatusCreated, status)
	}

	if userAgent, ok := logEntry["user_agent"].(string); !ok || userAgent != "test-agent" {
		t.Errorf("Expected user_agent 'test-agent', got %v", userAgent)
	}

	if _, ok := logEntry["duration_ms"]; !ok {
		t.Error("Expected duration field in log output")
	}
}

func TestResponseWriterWriteHeader(t *testing.T) {
	// Create a mock ResponseWriter
	mockRW := httptest.NewRecorder()

	// Create our responseWriter wrapper
	rw := &responseWriter{
		ResponseWriter: mockRW,
		status:         http.StatusOK, // Default status
	}

	// Call WriteHeader with a different status
	rw.WriteHeader(http.StatusNotFound)

	// Check that our wrapper captured the status
	if rw.status != http.StatusNotFound {
		t.Errorf("responseWriter did not capture status code: got %v want %v",
			rw.status, http.StatusNotFound)
	}

	// Check that the underlying ResponseWriter got the status
	if mockRW.Code != http.StatusNotFound {
		t.Errorf("underlying ResponseWriter did not get status code: got %v want %v",
			mockRW.Code, http.StatusNotFound)
	}
}

func TestResponseWriterWrite(t *testing.T) {
	// Create a mock ResponseWriter
	mockRW := httptest.NewRecorder()

	// Create our responseWriter wrapper
	rw := &responseWriter{
		ResponseWriter: mockRW,
		status:         http.StatusOK, // Default status
	}

	// Test data to write
	testData := []byte("test data")

	// Call Write method
	n, err := rw.Write(testData)

	// Check for errors
	if err != nil {
		t.Errorf("responseWriter.Write returned error: %v", err)
	}

	// Check that the correct number of bytes was reported as written
	if n != len(testData) {
		t.Errorf("responseWriter.Write returned wrong byte count: got %v want %v",
			n, len(testData))
	}

	// Check that the data was written to the underlying ResponseWriter
	if mockRW.Body.String() != string(testData) {
		t.Errorf("responseWriter did not write data correctly: got %v want %v",
			mockRW.Body.String(), string(testData))
	}

	// Check that status remains unchanged
	if rw.status != http.StatusOK {
		t.Errorf("responseWriter status changed unexpectedly: got %v want %v",
			rw.status, http.StatusOK)
	}
}

func TestStructuredLoggerWithDefaultStatus(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a JSON handler that writes to our buffer
	jsonHandler := slog.NewJSONHandler(&logBuffer, nil)
	oldDefault := slog.Default()
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(oldDefault) })

	// Create a test HTTP handler that doesn't explicitly set a status code
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("default status test"))
	})

	// Wrap our test handler with the StructuredLogger middleware
	middlewareHandler := WebRequestLogger(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/default-status", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	middlewareHandler.ServeHTTP(rr, req)

	// Check the response has default status 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse the log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(logBuffer.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify the status in the log is 200 OK
	if status, ok := logEntry["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Errorf("Expected status %d, got %v", http.StatusOK, status)
	}

	if id, ok := logEntry[requestKey].(string); !ok || id == "" {
		t.Errorf("Expected non-empty %q in log, got %v", requestKey, id)
	}

	// Verify remote address is logged
	if remote, ok := logEntry["remote"].(string); !ok || remote != "192.168.1.1:12345" {
		t.Errorf("Expected remote '192.168.1.1:12345', got %v", remote)
	}
}

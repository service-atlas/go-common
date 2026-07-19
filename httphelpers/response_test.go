package httphelpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/service-atlas/go-common/errorenvelope"
)

func TestWriteJSONResponse(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           any
		expectedStatus int
		expectedBody   string
		checkBody      bool
	}{
		{
			name:           "Success",
			status:         http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
			checkBody:      true,
		},
		{
			name:           "Created",
			status:         http.StatusCreated,
			data:           map[string]int{"id": 123},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":123}`,
			checkBody:      true,
		},
		{
			name:           "Encoding Error",
			status:         http.StatusOK,
			data:           make(chan int), // Channels cannot be marshaled to JSON
			expectedStatus: http.StatusInternalServerError,
			checkBody:      false, // HandleHttpError will write an error envelope
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			WriteJSONResponse(w, r, tt.status, tt.data)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkBody {
				var actual, expected any
				if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
					t.Fatalf("failed to unmarshal actual body: %v", err)
				}
				if err := json.Unmarshal([]byte(tt.expectedBody), &expected); err != nil {
					t.Fatalf("failed to unmarshal expected body: %v", err)
				}

				actualJSON, _ := json.Marshal(actual)
				expectedJSON, _ := json.Marshal(expected)

				if string(actualJSON) != string(expectedJSON) {
					t.Errorf("expected body %s, got %s", tt.expectedBody, w.Body.String())
				}

				if w.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
				}
			} else if tt.expectedStatus == http.StatusInternalServerError {
				if w.Header().Get("Content-Type") != "application/problem+json" {
					t.Errorf("expected Content-Type application/problem+json, got %s", w.Header().Get("Content-Type"))
				}

				var envelope errorenvelope.ErrorEnvelope
				if err := json.NewDecoder(w.Body).Decode(&envelope); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}

				expectedDetail := "Failed to encode response"
				if envelope.Detail != expectedDetail {
					t.Errorf("expected detail %q, got %q", expectedDetail, envelope.Detail)
				}
			}
		})
	}
}

package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHttpError(t *testing.T) {
	tests := []struct {
		name           string
		err            ErrorEnvelope
		statusCode     int
		expectedStatus int
		expectedBody   string
		expectedTitle  string
		expectedType   string
	}{
		{
			name:           "Empty Envelope",
			err:            ErrorEnvelope{},
			statusCode:     http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedTitle:  "Bad Request",
			expectedType:   "about:blank",
		},
		{
			name: "With Detail",
			err: ErrorEnvelope{
				Detail: "Validation failed",
			},
			statusCode:     http.StatusUnprocessableEntity,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedTitle:  "Unprocessable Entity",
			expectedType:   "about:blank",
		},
		{
			name: "Custom Type and Title",
			err: ErrorEnvelope{
				Type:  "https://example.com/probs/out-of-credit",
				Title: "You do not have enough credit.",
			},
			statusCode:     http.StatusForbidden,
			expectedStatus: http.StatusForbidden,
			expectedTitle:  "You do not have enough credit.",
			expectedType:   "https://example.com/probs/out-of-credit",
		},
		{
			name:           "Zero Status Code Defaults to 500",
			err:            ErrorEnvelope{Detail: "Something went wrong"},
			statusCode:     0,
			expectedStatus: http.StatusInternalServerError,
			expectedTitle:  "Internal Server Error",
			expectedType:   "about:blank",
		},
		{
			name:           "Negative Status Code Defaults to 500",
			err:            ErrorEnvelope{Detail: "Something went wrong"},
			statusCode:     -1,
			expectedStatus: http.StatusInternalServerError,
			expectedTitle:  "Internal Server Error",
			expectedType:   "about:blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			HandleHttpError(rr, tt.err, tt.statusCode)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/problem+json" {
				t.Errorf("expected Content-Type application/problem+json, got %q", contentType)
			}

			var got ErrorEnvelope
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if got.Status != tt.expectedStatus {
				t.Errorf("expected envelope status %d, got %d", tt.expectedStatus, got.Status)
			}

			if got.Title != tt.expectedTitle {
				t.Errorf("expected envelope title %q, got %q", tt.expectedTitle, got.Title)
			}

			if got.Type != tt.expectedType {
				t.Errorf("expected envelope type %q, got %q", tt.expectedType, got.Type)
			}
		})
	}
}

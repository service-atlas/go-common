package errorenvelope

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorEnvelope struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// HandleHttpError handles an error and writes it to the response writer
// in the format specified by RFC 7807. If an invalid status code is provided,
// it will default to 500 Internal Server Error.
func HandleHttpError(w http.ResponseWriter, err ErrorEnvelope, statusCode int) {
	if err.Type == "" {
		err.Type = "about:blank"
	}
	if http.StatusText(statusCode) == "" {
		slog.Error("Unknown HTTP status code", "status_code", statusCode)
		statusCode = http.StatusInternalServerError
	}
	if err.Title == "" {
		err.Title = http.StatusText(statusCode)
	}
	err.Status = statusCode

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(statusCode)

	if encodeErr := json.NewEncoder(w).Encode(err); encodeErr != nil {
		slog.Error("Failed to encode error response", "error", encodeErr)
	}
}

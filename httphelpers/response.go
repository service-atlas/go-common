package httphelpers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/service-atlas/go-common/errorenvelope"
	"github.com/service-atlas/go-common/httplog"
)

// WriteJSONResponse encodes the data to JSON and writes it to the response writer.
// It uses a buffer to ensure that encoding is successful before writing the header.
func WriteJSONResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to encode response"}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		httplog.LoggerFromContext(r.Context()).Error("Failed to write response", "request", r.URL.Path, "error", err)
	}
}

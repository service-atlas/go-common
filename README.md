# go-common
![Coverage](https://img.shields.io/badge/Coverage-94.4%25-brightgreen)

`go-common` is a utility module for Go services in the Service Atlas ecosystem, providing commonly used code for HTTP handling, logging, error reporting, and configuration.

## Packages

### `corsconfig`

Utilities for managing Cross-Origin Resource Sharing (CORS) configurations.

- **`GetCORSConfig() CORSConfig`**: Retrieves CORS configuration from the `cors_config` environment variable (JSON format). Returns a default configuration if the variable is missing or invalid.
- **`CORSConfig`**: A struct representing CORS settings including allowed origins, methods, headers, and credentials support.

### `errorenvelope`

Implements RFC 7807 (Problem Details for HTTP APIs) for standardized error responses.

- **`HandleHttpError(w http.ResponseWriter, err ErrorEnvelope, statusCode int)`**: Writes a standardized error response in `application/problem+json` format.
- **`ErrorEnvelope`**: A struct representing the error details according to RFC 7807.

### `httphelpers`

A collection of helpers for common HTTP tasks like response writing and path variable validation.

- **`WriteJSONResponse(w http.ResponseWriter, r *http.Request, status int, data any)`**: Encodes data as JSON and writes it to the response. It ensures encoding success before sending any headers.
- **`SetPathValueLookup(fn func(*http.Request, string) string)`**: Sets a custom lookup function for path variables (useful for routers that don't use standard Go 1.22+ path values).
- **`GetGuidFromRequestPath(varName string, req *http.Request) (string, bool)`**: Retrieves and validates a UUID from the request path.
- **`IsValidGuid(guidVal string) (string, bool)`**: Checks if a string is a valid UUID.
- **`GetDateFromRequestPath(varName string, req *http.Request) (time.Time, bool)`**: Retrieves and validates a date (YYYY-MM-DD) from the request path.
- **`GetIntFromRequestPath(varName string, req *http.Request) (int, bool)`**: Retrieves and validates a positive integer from the request path.

### `httplog`

Middleware and helpers for structured HTTP logging using `log/slog`.

- **`WebRequestLogger(next http.Handler) http.Handler`**: Middleware that logs HTTP requests, generates/propagates request IDs, and injects a structured logger into the request context.
- **`LoggerFromContext(ctx context.Context) *slog.Logger`**: Retrieves the contextual logger, falling back to `slog.Default()` if not found.
- **`GetRequestId(r *http.Request) string`**: Retrieves the request ID from the request context.
- **`GetRequestIdFromContext(ctx context.Context) string`**: Retrieves the request ID from a context.

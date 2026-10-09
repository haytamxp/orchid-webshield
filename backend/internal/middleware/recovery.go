package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type recoveryError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

type recoveryErrorEnvelope struct {
	Error recoveryError `json:"error"`
}

// Recover prevents a downstream handler panic from crashing the entire
// HTTP server process. Internal diagnostics are logged, not returned to clients.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}

				// Preserve net/http's special mechanism for aborting a request.
				if abortErr, ok := recovered.(error); ok &&
					abortErr == http.ErrAbortHandler {
					panic(recovered)
				}

				requestID := RequestIDFromContext(r.Context())

				logger.ErrorContext(
					r.Context(),
					"panic_recovered",
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
					"panic_type", fmt.Sprintf("%T", recovered),
					"stack", string(debug.Stack()),
				)

				w.Header().Set(
					"Content-Type",
					"application/json; charset=utf-8",
				)
				w.Header().Set("Cache-Control", "no-store")
				w.WriteHeader(http.StatusInternalServerError)

				_ = json.NewEncoder(w).Encode(
					recoveryErrorEnvelope{
						Error: recoveryError{
							Code:      "internal_server_error",
							Message:   "An unexpected error occurred.",
							RequestID: requestID,
						},
					},
				)
			}()

			next.ServeHTTP(w, r)
		})
	}
}

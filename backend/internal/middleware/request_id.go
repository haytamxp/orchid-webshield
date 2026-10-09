package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
)

const RequestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

type requestIDError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type requestIDErrorEnvelope struct {
	Error requestIDError `json:"error"`
}

// RequestID generates a server-controlled request identifier for each request.
//
// Any client-supplied X-Request-ID is overwritten. Request IDs are for
// correlation only and must never be treated as authentication credentials.
func RequestID(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := generateRequestID()
			if err != nil {
				logger.Error(
					"request_id_generation_failed",
					"error",
					err,
				)

				w.Header().Set(
					"Content-Type",
					"application/json; charset=utf-8",
				)
				w.Header().Set("Cache-Control", "no-store")
				w.WriteHeader(http.StatusInternalServerError)

				_ = json.NewEncoder(w).Encode(
					requestIDErrorEnvelope{
						Error: requestIDError{
							Code:    "internal_server_error",
							Message: "The request could not be processed.",
						},
					},
				)
				return
			}

			w.Header().Set(RequestIDHeader, id)

			ctx := context.WithValue(
				r.Context(),
				requestIDContextKey{},
				id,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext retrieves the request ID assigned by RequestID.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey{}).(string)
	return id
}

func generateRequestID() (string, error) {
	var randomBytes [16]byte

	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes[:]), nil
}

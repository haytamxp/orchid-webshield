package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/haytamxp/orchid-webshield/backend/internal/handlers"
	"github.com/haytamxp/orchid-webshield/backend/internal/middleware"
)

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

// NewRouter creates the HTTP router and wraps it with cross-cutting controls.
//
// maxRequestBodyBytes must come from validated configuration.
func NewRouter(
	logger *slog.Logger,
	maxRequestBodyBytes int64,
) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()

	mux.Handle(
		"/api/v1/health",
		handlers.NewHealthHandler(),
	)

	// This explicit fallback returns consistent JSON for unknown paths.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeAPIError(
			w,
			r,
			http.StatusNotFound,
			"not_found",
			"The requested resource was not found.",
		)
	})

	var handler http.Handler = mux

	handler = limitRequestBody(maxRequestBodyBytes, handler)
	handler = middleware.Recover(logger)(handler)
	handler = middleware.RequestID(logger)(handler)

	return securityHeaders(handler)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'none'; frame-ancestors 'none'",
		)
		w.Header().Set(
			"Permissions-Policy",
			"camera=(), microphone=(), geolocation=()",
		)

		next.ServeHTTP(w, r)
	})
}

// limitRequestBody bounds request bodies before they reach application
// handlers. Known oversized Content-Length values are rejected immediately.
// Unknown-length bodies are read through a bounded reader before dispatch.
func limitRequestBody(
	maxBytes int64,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if maxBytes < 1 {
			writeAPIError(
				w,
				r,
				http.StatusInternalServerError,
				"server_configuration_error",
				"The server could not process the request.",
			)
			return
		}

		if r.Body == nil || r.Body == http.NoBody {
			next.ServeHTTP(w, r)
			return
		}

		if r.ContentLength > maxBytes {
			_ = r.Body.Close()

			writeAPIError(
				w,
				r,
				http.StatusRequestEntityTooLarge,
				"request_body_too_large",
				"The request body exceeds the allowed size.",
			)
			return
		}

		if r.ContentLength >= 0 {
			// The declared body length is within the configured limit.
			// MaxBytesReader also protects handlers that read the body.
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
			return
		}

		// ContentLength < 0 means the request length is unknown.
		// Buffer no more than maxBytes + 1 to detect oversized bodies.
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes+1))
		_ = r.Body.Close()

		if err != nil {
			writeAPIError(
				w,
				r,
				http.StatusBadRequest,
				"invalid_request_body",
				"The request body could not be read.",
			)
			return
		}

		if int64(len(body)) > maxBytes {
			writeAPIError(
				w,
				r,
				http.StatusRequestEntityTooLarge,
				"request_body_too_large",
				"The request body exceeds the allowed size.",
			)
			return
		}

		// Restore the bounded body so that the downstream handler can read it.
		r.Body = io.NopCloser(bytes.NewReader(body))

		next.ServeHTTP(w, r)
	})
}

func writeAPIError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code string,
	message string,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)

	response := apiErrorEnvelope{
		Error: apiError{
			Code:      code,
			Message:   message,
			RequestID: middleware.RequestIDFromContext(r.Context()),
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}

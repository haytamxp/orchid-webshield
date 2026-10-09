package handlers

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type healthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type healthErrorEnvelope struct {
	Error healthError `json:"error"`
}

// NewHealthHandler returns a minimal liveness endpoint.
//
// This endpoint intentionally checks the HTTP application itself only.
// It does not yet establish database or worker readiness.
func NewHealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")

			writeJSON(
				w,
				http.StatusMethodNotAllowed,
				healthErrorEnvelope{
					Error: healthError{
						Code:    "method_not_allowed",
						Message: http.StatusText(http.StatusMethodNotAllowed),
					},
				},
			)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)

		// HEAD returns the same status and headers as GET but no body.
		if r.Method == http.MethodHead {
			return
		}

		response := healthResponse{
			Status:  "ok",
			Service: "orchid-webshield-backend",
		}

		_ = json.NewEncoder(w).Encode(response)
	})
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	payload any,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

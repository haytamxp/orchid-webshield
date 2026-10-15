package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haytamxp/orchid-webshield/webapp/internal/securityevents"
)

func TestLoginFailureWritesSecurityEvent(t *testing.T) {
	db := newHandlerTestDB(t)
	logPath := filepath.Join(t.TempDir(), "security_events.jsonl")

	h := &Handler{
		DB:     db,
		Events: securityevents.New(logPath),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", h.Login)

	rec := doRequest(
		mux,
		http.MethodPost,
		"/login",
		`{"username":"unknown-user","password":"wrong-password"}`,
	)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusUnauthorized,
		)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("security event file was not created: %v", err)
	}

	if !strings.Contains(string(data), `"event_type":"login_failure"`) {
		t.Fatalf("login_failure event missing from log: %s", data)
	}
}
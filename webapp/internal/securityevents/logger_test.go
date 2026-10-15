package securityevents

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLogRequestCreatesJSONLFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "logs", "security_events.jsonl")
	logger := New(logPath)

	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	err := logger.LogRequest(
		req,
		"login_failure",
		"MEDIUM",
		nil,
		401,
		map[string]any{"reason": "invalid_credentials"},
	)
	if err != nil {
		t.Fatalf("LogRequest failed: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("cannot read JSONL file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("JSONL file is empty")
	}
}
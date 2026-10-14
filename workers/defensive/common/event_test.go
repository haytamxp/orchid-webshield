package common

import (
	"encoding/json"
	"testing"
	"time"
)

func validTestEvent() SecurityEvent {
	return SecurityEvent{
		SchemaVersion: 1,
		EventID:       "evt-test-001",
		Timestamp:     time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC),
		Source:        "webapp",
		EventType:     "authorization_denied",
		Severity:      "medium",
		SrcIP:         "192.168.1.10",
		RequestID:     "req-test-001",
	}
}

func TestSecurityEventJSON(t *testing.T) {
	user := "user01"
	statusCode := 403

	event := validTestEvent()
	event.User = &user
	event.Endpoint = "/api/orders/123"
	event.Method = "GET"
	event.StatusCode = &statusCode
	event.UserAgent = "curl/8.4"
	event.Details = map[string]any{
		"reason": "not_owner",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON decoding failed: %v", err)
	}

	expected := map[string]any{
		"schema_version": float64(1),
		"event_id":       "evt-test-001",
		"timestamp":      "2026-10-13T10:00:00Z",
		"source":         "webapp",
		"event_type":     "authorization_denied",
		"severity":       "medium",
		"src_ip":         "192.168.1.10",
		"request_id":     "req-test-001",
		"user":           "user01",
		"endpoint":       "/api/orders/123",
		"method":         "GET",
		"status_code":    float64(403),
		"user_agent":     "curl/8.4",
	}

	for key, want := range expected {
		if decoded[key] != want {
			t.Errorf("%s: got %v, want %v", key, decoded[key], want)
		}
	}

	details, ok := decoded["details"].(map[string]any)
	if !ok {
		t.Fatalf("details should be a JSON object, got %T", decoded["details"])
	}

	if details["reason"] != "not_owner" {
		t.Errorf("unexpected details.reason: %v", details["reason"])
	}

	if _, exists := decoded["source_ip"]; exists {
		t.Error("source_ip should not be present; contract uses src_ip")
	}
}

func TestSecurityEventOptionalFields(t *testing.T) {
	event := validTestEvent()

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON decoding failed: %v", err)
	}

	optionalFields := []string{
		"actor_id",
		"user",
		"target",
		"endpoint",
		"method",
		"status_code",
		"outcome",
		"reason_code",
		"user_agent",
		"details",
	}

	for _, field := range optionalFields {
		if _, exists := decoded[field]; exists {
			t.Errorf("%s should be omitted when empty", field)
		}
	}
}

func TestSecurityEventValidate(t *testing.T) {
	event := validTestEvent()

	if err := event.Validate(); err != nil {
		t.Errorf("valid event should pass validation: %v", err)
	}
}

func TestSecurityEventValidateRejectsInvalidEvents(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*SecurityEvent)
	}{
		{
			name: "invalid schema version",
			modify: func(e *SecurityEvent) {
				e.SchemaVersion = 2
			},
		},
		{
			name: "missing event ID",
			modify: func(e *SecurityEvent) {
				e.EventID = ""
			},
		},
		{
			name: "missing timestamp",
			modify: func(e *SecurityEvent) {
				e.Timestamp = time.Time{}
			},
		},
		{
			name: "missing source",
			modify: func(e *SecurityEvent) {
				e.Source = ""
			},
		},
		{
			name: "missing event type",
			modify: func(e *SecurityEvent) {
				e.EventType = ""
			},
		},
		{
			name: "missing request ID",
			modify: func(e *SecurityEvent) {
				e.RequestID = ""
			},
		},
		{
			name: "missing source IP",
			modify: func(e *SecurityEvent) {
				e.SrcIP = ""
			},
		},
		{
			name: "invalid source IP",
			modify: func(e *SecurityEvent) {
				e.SrcIP = "bonjour"
			},
		},
		{
			name: "invalid severity",
			modify: func(e *SecurityEvent) {
				e.Severity = "BANANA"
			},
		},
		{
			name: "uppercase severity",
			modify: func(e *SecurityEvent) {
				e.Severity = "MEDIUM"
			},
		},
		{
			name: "non-UTC timestamp",
			modify: func(e *SecurityEvent) {
				e.Timestamp = time.Date(
					2026, 10, 13, 10, 0, 0, 0,
					time.FixedZone("UTC+1", 3600),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := validTestEvent()
			tt.modify(&event)

			if err := event.Validate(); err == nil {
				t.Error("invalid event should fail validation")
			}
		})
	}
}

func TestSecurityEventValidateAcceptsIPv4AndIPv6(t *testing.T) {
	ips := []string{
		"192.168.1.10",
		"2001:db8::1",
	}

	for _, ip := range ips {
		t.Run(ip, func(t *testing.T) {
			event := validTestEvent()
			event.SrcIP = ip

			if err := event.Validate(); err != nil {
				t.Errorf("valid IP %q should pass validation: %v", ip, err)
			}
		})
	}
}

func TestSecurityEventValidateAcceptsSupportedSeverities(t *testing.T) {
	severities := []string{
		"info",
		"low",
		"medium",
		"high",
		"critical",
	}

	for _, severity := range severities {
		t.Run(severity, func(t *testing.T) {
			event := validTestEvent()
			event.Severity = severity

			if err := event.Validate(); err != nil {
				t.Errorf("severity %q should be accepted: %v", severity, err)
			}
		})
	}
}

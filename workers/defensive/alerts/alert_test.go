
package alerts

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/haytamxp/orchid-webshield/workers/defensive/detection"
)

func TestFromDetection(t *testing.T) {
	detectedAt := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	d := detection.Detection{
		RuleID:     "WEB-BF-001",
		Title:      "Multiple failed login attempts from one IP",
		Severity:   "high",
		SourceIP:   "192.168.1.10",
		EventIDs:   []string{"evt-1", "evt-2", "evt-3", "evt-4", "evt-5"},
		DetectedAt: detectedAt,
	}

	alert := FromDetection(d, "alert-001")

	if alert.SchemaVersion != 1 {
		t.Errorf("expected schema version 1, got %d", alert.SchemaVersion)
	}
	if alert.AlertID != "alert-001" {
		t.Errorf("unexpected alert ID: %s", alert.AlertID)
	}
	if alert.Source != "defensive-worker" {
		t.Errorf("unexpected source: %s", alert.Source)
	}
	if alert.RuleID != d.RuleID {
		t.Errorf("unexpected rule ID: %s", alert.RuleID)
	}
	if alert.Severity != "high" {
		t.Errorf("unexpected severity: %s", alert.Severity)
	}
	if len(alert.EventIDs) != len(d.EventIDs) {
		t.Fatalf("expected %d event IDs, got %d", len(d.EventIDs), len(alert.EventIDs))
	}
	if alert.CreatedAt.IsZero() {
		t.Fatal("created_at must not be zero")
	}
	if alert.CreatedAt.Location() != time.UTC {
		t.Errorf("expected UTC creation time, got %v", alert.CreatedAt.Location())
	}
}

func TestAlertJSONContract(t *testing.T) {
	alert := Alert{
		SchemaVersion: 1,
		AlertID:       "alert-001",
		CreatedAt:     time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
		Source:        "defensive-worker",
		RuleID:        "WEB-BF-001",
		Title:         "Multiple failed login attempts from one IP",
		Severity:      "high",
		EventIDs:      []string{"evt-1", "evt-2"},
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("failed to marshal alert: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode alert JSON: %v", err)
	}

	requiredFields := []string{
		"schema_version",
		"alert_id",
		"created_at",
		"source",
		"rule_id",
		"title",
		"severity",
		"event_ids",
	}

	for _, field := range requiredFields {
		if _, exists := decoded[field]; !exists {
			t.Errorf("missing required JSON field: %s", field)
		}
	}

	if decoded["source"] != "defensive-worker" {
		t.Errorf("unexpected source: %v", decoded["source"])
	}

	if decoded["rule_id"] != "WEB-BF-001" {
		t.Errorf("unexpected rule_id: %v", decoded["rule_id"])
	}
}

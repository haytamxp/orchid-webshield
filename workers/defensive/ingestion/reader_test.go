package ingestion

import (
	"strings"
	"testing"
)

func TestReadEventsAcceptsValidEvents(t *testing.T) {
	input := `{"schema_version":1,"event_id":"evt-001","timestamp":"2026-10-13T10:00:00Z","source":"webapp","event_type":"login_success","src_ip":"192.168.1.10","severity":"info","request_id":"req-001"}
{"schema_version":1,"event_id":"evt-002","timestamp":"2026-10-13T10:01:00Z","source":"webapp","event_type":"login_failure","src_ip":"192.168.1.11","severity":"low","request_id":"req-002"}`

	events, errs := ReadEvents(strings.NewReader(input))

	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].EventType != "login_success" {
		t.Errorf("unexpected first event type: %s", events[0].EventType)
	}

	if events[1].EventID != "evt-002" {
		t.Errorf("unexpected second event ID: %s", events[1].EventID)
	}
}

func TestReadEventsReportsInvalidJSON(t *testing.T) {
	input := `{"timestamp":`

	events, errs := ReadEvents(strings.NewReader(input))

	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}

	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if !strings.Contains(errs[0].Error(), "line 1") {
		t.Errorf("expected error to mention line 1, got: %v", errs[0])
	}
}

func TestReadEventsRejectsInvalidEvent(t *testing.T) {
	input := `{"schema_version":1,"event_id":"evt-003","timestamp":"2026-10-13T10:00:00Z","source":"webapp","event_type":"login_success","severity":"info","request_id":"req-003"}`

	events, errs := ReadEvents(strings.NewReader(input))

	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}

	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if !strings.Contains(errs[0].Error(), "src_ip") {
		t.Errorf("expected missing or invalid src_ip error, got: %v", errs[0])
	}
}

func TestReadEventsContinuesAfterInvalidLine(t *testing.T) {
	input := `{"schema_version":1,"event_id":"evt-004","timestamp":"2026-10-13T10:00:00Z","source":"webapp","event_type":"login_success","src_ip":"192.168.1.10","severity":"info","request_id":"req-004"}
not-valid-json
{"schema_version":1,"event_id":"evt-005","timestamp":"2026-10-13T10:02:00Z","source":"webapp","event_type":"logout","src_ip":"192.168.1.12","severity":"low","request_id":"req-005"}`

	events, errs := ReadEvents(strings.NewReader(input))

	if len(events) != 2 {
		t.Errorf("expected 2 valid events, got %d", len(events))
	}

	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(errs), errs)
	}

	if len(events) == 2 && events[1].EventID != "evt-005" {
		t.Errorf("expected the second valid event to be evt-005, got %s", events[1].EventID)
	}
}

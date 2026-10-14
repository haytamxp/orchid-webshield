
package detection

import (
	"fmt"
	"testing"
	"time"

	"github.com/haytamxp/orchid-webshield/workers/defensive/common"
)

func makeLoginFailure(id string, timestamp time.Time, ip string) common.SecurityEvent {
	return common.SecurityEvent{
		SchemaVersion: 1,
		EventID:       id,
		Timestamp:     timestamp,
		Source:        "webapp",
		EventType:     "login_failure",
		Severity:      "low",
		SrcIP:         ip,
		RequestID:     fmt.Sprintf("req-%s", id),
	}
}

func TestDetectBruteForce(t *testing.T) {
	engine := NewEngine()
	start := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	var events []common.SecurityEvent
	for i := 0; i < 5; i++ {
		events = append(events, makeLoginFailure(
			fmt.Sprintf("evt-%d", i),
			start.Add(time.Duration(i)*20*time.Second),
			"192.168.1.10",
		))
	}

	detections := engine.DetectBruteForce(events)

	if len(detections) != 1 {
		t.Fatalf("expected 1 detection, got %d", len(detections))
	}

	if detections[0].RuleID != "WEB-BF-001" {
		t.Errorf("unexpected rule ID: %s", detections[0].RuleID)
	}

	if len(detections[0].EventIDs) != 5 {
		t.Errorf("expected 5 event IDs, got %d", len(detections[0].EventIDs))
	}
}

func TestDetectBruteForceIgnoresSuccessfulLogins(t *testing.T) {
	engine := NewEngine()
	start := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	var events []common.SecurityEvent
	for i := 0; i < 5; i++ {
		event := makeLoginFailure(
			fmt.Sprintf("evt-%d", i),
			start.Add(time.Duration(i)*10*time.Second),
			"192.168.1.10",
		)
		if i == 0 {
			event.EventType = "login_success"
		}
		events = append(events, event)
	}

	detections := engine.DetectBruteForce(events)

	if len(detections) != 0 {
		t.Fatalf("expected no detection, got %d", len(detections))
	}
}

func TestDetectBruteForceRequiresThreshold(t *testing.T) {
	engine := NewEngine()
	start := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	events := []common.SecurityEvent{
		makeLoginFailure("evt-1", start, "192.168.1.10"),
		makeLoginFailure("evt-2", start.Add(30*time.Second), "192.168.1.10"),
		makeLoginFailure("evt-3", start.Add(time.Minute), "192.168.1.10"),
		makeLoginFailure("evt-4", start.Add(90*time.Second), "192.168.1.10"),
	}

	detections := engine.DetectBruteForce(events)

	if len(detections) != 0 {
		t.Fatalf("expected no detection below threshold, got %d", len(detections))
	}
}

func TestDetectBruteForceOutsideWindow(t *testing.T) {
	engine := NewEngine()
	start := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	var events []common.SecurityEvent
	for i := 0; i < 5; i++ {
		events = append(events, makeLoginFailure(
			fmt.Sprintf("evt-%d", i),
			start.Add(time.Duration(i)*40*time.Second),
			"192.168.1.10",
		))
	}

	detections := engine.DetectBruteForce(events)

	if len(detections) != 0 {
		t.Fatalf("expected no detection outside the time window, got %d", len(detections))
	}
}

func TestDetectBruteForceOutOfOrderEvents(t *testing.T) {
	engine := NewEngine()
	start := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	events := []common.SecurityEvent{
		makeLoginFailure("evt-5", start.Add(80*time.Second), "192.168.1.10"),
		makeLoginFailure("evt-2", start.Add(20*time.Second), "192.168.1.10"),
		makeLoginFailure("evt-4", start.Add(60*time.Second), "192.168.1.10"),
		makeLoginFailure("evt-1", start, "192.168.1.10"),
		makeLoginFailure("evt-3", start.Add(40*time.Second), "192.168.1.10"),
	}

	detections := engine.DetectBruteForce(events)

	if len(detections) != 1 {
		t.Fatalf("expected 1 detection, got %d", len(detections))
	}

	if len(detections[0].EventIDs) != 5 {
		t.Fatalf("expected 5 event IDs, got %d", len(detections[0].EventIDs))
	}
}

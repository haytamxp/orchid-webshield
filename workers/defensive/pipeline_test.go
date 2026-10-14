
package defensive

import (
	"fmt"
	"testing"
	"time"

	"github.com/haytamxp/orchid-webshield/workers/defensive/common"
)

func TestProcessEventsGeneratesAlert(t *testing.T) {
	start := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

	var events []common.SecurityEvent
	for i := 0; i < 5; i++ {
		events = append(events, common.SecurityEvent{
			SchemaVersion: 1,
			EventID:       fmt.Sprintf("evt-%d", i),
			Timestamp:     start.Add(time.Duration(i)*20*time.Second),
			Source:        "webapp",
			EventType:     "login_failure",
			Severity:      "low",
			SrcIP:         "192.168.1.10",
			RequestID:     fmt.Sprintf("req-%d", i),
		})
	}

	got := ProcessEvents(events)

	if len(got) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(got))
	}

	if got[0].RuleID != "WEB-BF-001" {
		t.Errorf("unexpected rule ID: %s", got[0].RuleID)
	}

	if len(got[0].EventIDs) != 5 {
		t.Errorf("expected 5 event IDs, got %d", len(got[0].EventIDs))
	}
}

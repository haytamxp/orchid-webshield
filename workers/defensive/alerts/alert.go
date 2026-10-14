
package alerts

import (
	"time"

	"github.com/haytamxp/orchid-webshield/workers/defensive/detection"
)

// Alert represents a security alert generated from a detection.
type Alert struct {
	SchemaVersion int       `json:"schema_version"`
	AlertID        string    `json:"alert_id"`
	CreatedAt      time.Time `json:"created_at"`
	Source         string    `json:"source"`
	RuleID         string    `json:"rule_id"`
	Title          string    `json:"title"`
	Severity       string    `json:"severity"`
	EventIDs       []string  `json:"event_ids"`
}

// FromDetection converts a detection into an alert.
func FromDetection(d detection.Detection, alertID string) Alert {
	return Alert{
		SchemaVersion: 1,
		AlertID:        alertID,
		CreatedAt:      time.Now().UTC(),
		Source:         "defensive-worker",
		RuleID:         d.RuleID,
		Title:          d.Title,
		Severity:       d.Severity,
		EventIDs:       append([]string(nil), d.EventIDs...),
	}
}

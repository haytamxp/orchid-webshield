package common

import "time"

// SecurityEvent represents a normalized security event.
type SecurityEvent struct {
	SchemaVersion int       `json:"schema_version"`
	EventID       string    `json:"event_id"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
	EventType     string    `json:"event_type"`
	Severity      string    `json:"severity"`
	SrcIP         string    `json:"src_ip"`
	RequestID     string    `json:"request_id"`

	ActorID    string         `json:"actor_id,omitempty"`
	User       *string        `json:"user,omitempty"`
	Target     string         `json:"target,omitempty"`
	Endpoint   string         `json:"endpoint,omitempty"`
	Method     string         `json:"method,omitempty"`
	StatusCode *int           `json:"status_code,omitempty"`
	Outcome    string         `json:"outcome,omitempty"`
	ReasonCode string         `json:"reason_code,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

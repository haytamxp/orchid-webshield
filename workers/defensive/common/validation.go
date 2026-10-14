package common

import (
	"fmt"
	"net"
	"strings"
)

// Validate checks whether the event respects the WebShield contract.
func (e SecurityEvent) Validate() error {
	if e.SchemaVersion != 1 {
		return fmt.Errorf("schema_version must be 1")
	}

	if strings.TrimSpace(e.EventID) == "" {
		return fmt.Errorf("event_id is required")
	}

	if e.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	_, offset := e.Timestamp.Zone()
	if offset != 0 {
		return fmt.Errorf("timestamp must use UTC")
	}

	if strings.TrimSpace(e.Source) == "" {
		return fmt.Errorf("source is required")
	}

	if strings.TrimSpace(e.EventType) == "" {
		return fmt.Errorf("event_type is required")
	}

	if strings.TrimSpace(e.RequestID) == "" {
		return fmt.Errorf("request_id is required")
	}

	if net.ParseIP(e.SrcIP) == nil {
		return fmt.Errorf("src_ip must be a valid IP address")
	}

	switch e.Severity {
	case "info", "low", "medium", "high", "critical":
		// Severity is valid.
	default:
		return fmt.Errorf("severity must be info, low, medium, high, or critical")
	}

	return nil
}

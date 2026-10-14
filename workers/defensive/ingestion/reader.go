
package ingestion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/haytamxp/orchid-webshield/workers/defensive/common"
)

// ReadEvents reads JSON Lines and returns valid security events.
// Invalid lines are reported without stopping the entire ingestion.
func ReadEvents(r io.Reader) ([]common.SecurityEvent, []error) {
	scanner := bufio.NewScanner(r)

	// Allow JSON lines larger than the default scanner limit.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var events []common.SecurityEvent
	var errors []error
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()

		if len(line) == 0 {
			continue
		}

		var event common.SecurityEvent
		if err := json.Unmarshal(line, &event); err != nil {
			errors = append(errors, fmt.Errorf(
				"line %d: invalid JSON: %w", lineNumber, err,
			))
			continue
		}

		if err := event.Validate(); err != nil {
			errors = append(errors, fmt.Errorf(
				"line %d: invalid event: %w", lineNumber, err,
			))
			continue
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		errors = append(errors, fmt.Errorf("reading input: %w", err))
	}

	return events, errors
}



package defensive

import (
	"fmt"

	"github.com/haytamxp/orchid-webshield/workers/defensive/alerts"
	"github.com/haytamxp/orchid-webshield/workers/defensive/common"
	"github.com/haytamxp/orchid-webshield/workers/defensive/detection"
)

// ProcessEvents analyzes security events and returns generated alerts.
func ProcessEvents(events []common.SecurityEvent) []alerts.Alert {
	engine := detection.NewEngine()
	detections := engine.DetectBruteForce(events)

	result := make([]alerts.Alert, 0, len(detections))

	for i, detected := range detections {
		alertID := fmt.Sprintf("alert-%s-%d", detected.RuleID, i+1)
		result = append(result, alerts.FromDetection(detected, alertID))
	}

	return result
}

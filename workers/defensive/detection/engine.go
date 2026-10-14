
package detection

import (
	"time"

	"github.com/haytamxp/orchid-webshield/workers/defensive/common"
)
import "sort"
// Detection represents a suspicious pattern detected in security events.
type Detection struct {
	RuleID    string
	Title     string
	Severity  string
	SourceIP  string
	EventIDs  []string
	DetectedAt time.Time
}

// Engine evaluates normalized security events.
type Engine struct {
	BruteForceThreshold int
	BruteForceWindow    time.Duration
}

// NewEngine creates a detection engine with default brute-force settings.
func NewEngine() *Engine {
	return &Engine{
		BruteForceThreshold: 5,
		BruteForceWindow:    2 * time.Minute,
	}
}

// DetectBruteForce detects repeated failed logins from the same source IP.


func (e *Engine) DetectBruteForce(events []common.SecurityEvent) []Detection {
	if e.BruteForceThreshold <= 0 || e.BruteForceWindow <= 0 {
		return nil
	}

	// Regrouper les échecs de connexion par adresse IP.
	failuresByIP := make(map[string][]common.SecurityEvent)

	for _, event := range events {
		if event.EventType != "login_failure" {
			continue
		}

		if event.SrcIP == "" || event.Timestamp.IsZero() {
			continue
		}

		failuresByIP[event.SrcIP] = append(
			failuresByIP[event.SrcIP],
			event,
		)
	}

	var detections []Detection

	for ip, failures := range failuresByIP {
		// Trier les événements par ordre chronologique.
		sort.Slice(failures, func(i, j int) bool {
			return failures[i].Timestamp.Before(failures[j].Timestamp)
		})

		// Examiner chaque fenêtre temporelle.
		for start := 0; start < len(failures); start++ {
			end := start

			for end < len(failures) &&
				failures[end].Timestamp.Sub(failures[start].Timestamp) <= e.BruteForceWindow {
				end++
			}

			if end-start < e.BruteForceThreshold {
				continue
			}

			eventIDs := make([]string, 0, end-start)
			for _, event := range failures[start:end] {
				eventIDs = append(eventIDs, event.EventID)
			}

			detections = append(detections, Detection{
				RuleID:     "WEB-BF-001",
				Title:      "Multiple failed login attempts from one IP",
				Severity:   "high",
				SourceIP:   ip,
				EventIDs:   eventIDs,
				DetectedAt: failures[end-1].Timestamp,
			})

			// Une seule détection par IP pour cette exécution.
			break
		}
	}

	return detections
}

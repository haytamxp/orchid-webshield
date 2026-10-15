
package securityevents

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Event struct {
	Timestamp  string         `json:"timestamp"`
	Source     string         `json:"source"`
	EventType  string         `json:"event_type"`
	SrcIP      string         `json:"src_ip"`
	Severity   string         `json:"severity"`
	User       *string        `json:"user"`
	Endpoint   string         `json:"endpoint,omitempty"`
	Method     string         `json:"method,omitempty"`
	StatusCode int            `json:"status_code,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

type Logger struct {
	Path string
	mu   sync.Mutex
}

func New(path string) *Logger {
	return &Logger{Path: path}
}

func (l *Logger) LogRequest(
	r *http.Request,
	eventType string,
	severity string,
	username *string,
	status int,
	details map[string]any,
) error {
	if l == nil || strings.TrimSpace(l.Path) == "" {
		return nil
	}

	srcIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		srcIP = host
	}

	event := Event{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Source:     "webapp",
		EventType:  eventType,
		SrcIP:      srcIP,
		Severity:   severity,
		User:       username,
		Endpoint:   r.URL.Path,
		Method:     r.Method,
		StatusCode: status,
		UserAgent:  r.UserAgent(),
		Details:    details,
	}

	line, err := json.Marshal(event)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(l.Path), 0750); err != nil {
		return err
	}

	file, err := os.OpenFile(l.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(line)
	return err
}

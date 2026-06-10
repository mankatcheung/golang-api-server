// Package logger provides structured logging utilities including a custom
// slog.Handler that publishes log records to Kafka for persistent storage.
package logger

import "time"

// LogRecord is the payload written to the Kafka app-logs topic.
type LogRecord struct {
	Time    time.Time      `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

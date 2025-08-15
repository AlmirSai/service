package logger

import (
	"context"
	"log/slog"
	"time"
)

// Level is a custom log level type based on slog.Level.
// It allows extending or customizing logging behavior while staying compatible with slog.
type Level slog.Level

// Common log levels wrapped into our custom Level type.
const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// Record represents a structured log entry.
// Unlike slog.Record, this struct can be easily stored, serialized, or sent over the network.
type Record struct {
	Time       time.Time      // Timestamp of the log event
	Message    string         // Main log message
	Level      Level          // Log level
	Attributes map[string]any // Additional structured attributes
}

// toRecord converts a slog.Record into our custom Record format.
// This makes it possible to store or process log data outside of slog.
func toRecord(r slog.Record) Record {
	attrs := make(map[string]any, r.NumAttrs()) // Pre-allocate attribute map

	// Iterate over all attributes and store them in the map
	f := func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	}
	r.Attrs(f)

	return Record{
		Time:       r.Time,
		Message:    r.Message,
		Level:      Level(r.Level),
		Attributes: attrs,
	}
}

// EventFn defines a function type for handling log events.
// It receives the context and the log record, enabling async or external processing.
type EventFn func(ctx context.Context, r Record)

// Events holds callbacks for different log levels.
// This allows assigning custom behavior for each log level independently.
type Events struct {
	Debug EventFn // Called for debug-level logs
	Info  EventFn // Called for info-level logs
	Warn  EventFn // Called for warning-level logs
	Error EventFn // Called for error-level logs
}

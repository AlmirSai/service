package logger

import (
	"context"
	"log/slog"
)

// logHandler is a wrapper around slog.Handler that adds custom event hooks.
// It allows executing additional logic (e.g., sending errors to Sentry) 
// while still passing logs to the original handler.
type logHandler struct {
	handler slog.Handler // The underlying slog handler
	events  Events       // Custom event handlers for different log levels
}

// newLogHandler creates a new logHandler wrapping an existing slog.Handler 
// with custom event hooks.
func newLogHandler(handler slog.Handler, events Events) *logHandler {
	return &logHandler{
		handler: handler,
		events:  events,
	}
}

// Enabled checks whether the given log level is enabled for this handler.
func (h *logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// WithAttrs returns a new handler with additional attributes attached.
// The custom events are preserved.
func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logHandler{
		handler: h.handler.WithAttrs(attrs),
		events:  h.events,
	}
}

// WithGroup returns a new handler that groups all attributes under the given name.
// The custom events are preserved.
func (h *logHandler) WithGroup(name string) slog.Handler {
	return &logHandler{
		handler: h.handler.WithGroup(name),
		events:  h.events,
	}
}

// Handle processes a log record:
// 1. Executes the corresponding custom event hook based on log level.
// 2. Passes the record to the underlying slog.Handler for normal processing.
func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	switch r.Level {
	case slog.LevelDebug:
		if h.events.Debug != nil {
			h.events.Debug(ctx, toRecord(r))
		}
	case slog.LevelError:
		if h.events.Error != nil {
			h.events.Error(ctx, toRecord(r))
		}
	case slog.LevelWarn:
		if h.events.Warn != nil {
			h.events.Warn(ctx, toRecord(r))
		}
	case slog.LevelInfo:
		if h.events.Info != nil {
			h.events.Info(ctx, toRecord(r))
		}
	}

	// Always pass the record to the original handler
	return h.handler.Handle(ctx, r)
}
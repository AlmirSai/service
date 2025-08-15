package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

// TraceIDFn defines a function type for extracting a trace ID from the context.
// Useful for correlating logs in distributed systems.
type TraceIDFn func(ctx context.Context) string

// Logger is a structured logging wrapper around slog.Handler.
// It supports trace ID injection, service name tagging, and custom event hooks.
type Logger struct {
	discard   bool         // Whether logs should be discarded (io.Discard)
	handler   slog.Handler // Underlying slog handler
	traceIDFn TraceIDFn    // Function to extract trace ID from context
}

// New creates a Logger with the given output, log level, service name, and optional trace ID function.
func New(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn) *Logger {
	return new(w, minLevel, serviceName, traceIDFn, Events{})
}

// NewWithEvents creates a Logger with custom event hooks for different log levels.
func NewWithEvents(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn, events Events) *Logger {
	return new(w, minLevel, serviceName, traceIDFn, events)
}

// NewWithHandler wraps an existing slog.Handler in a Logger.
func NewWithHandler(h slog.Handler) *Logger {
	return &Logger{
		handler: h,
	}
}

// NewStdLogger creates a standard library log.Logger using the underlying slog handler.
// Useful for compatibility with packages expecting the old log.Logger API.
func NewStdLogger(logger *Logger, level Level) *log.Logger {
	return slog.NewLogLogger(logger.handler, slog.Level(level))
}

// Debug logs a debug-level message.
func (log *Logger) Debug(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}
	log.write(ctx, LevelDebug, 3, msg, args...)
}

// Debugc logs a debug-level message with a custom caller skip depth.
func (log *Logger) Debugc(ctx context.Context, caller int, msg string, args ...any) {
	if log.discard {
		return
	}
	log.write(ctx, LevelDebug, caller, msg, args...)
}

// Info logs an info-level message.
func (log *Logger) Info(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}
	log.write(ctx, LevelInfo, 3, msg, args...)
}

// Infoc logs an info-level message with a custom caller skip depth.
func (log *Logger) Infoc(ctx context.Context, caller int, msg string, args ...any) {
	if log.discard {
		return
	}
	log.write(ctx, LevelInfo, caller, msg, args...)
}

// Warn logs a warning-level message.
func (log *Logger) Warn(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}
	log.write(ctx, LevelWarn, 3, msg, args...)
}

// Warnc logs a warning-level message with a custom caller skip depth.
func (log *Logger) Warnc(ctx context.Context, caller int, msg string, args ...any) {
	if log.discard {
		return
	}
	log.write(ctx, LevelWarn, caller, msg, args...)
}

// Error logs an error-level message.
func (log *Logger) Error(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}
	log.write(ctx, LevelError, 3, msg, args...)
}

// Errorc logs an error-level message with a custom caller skip depth.
func (log *Logger) Errorc(ctx context.Context, caller int, msg string, args ...any) {
	if log.discard {
		return
	}
	log.write(ctx, LevelError, caller, msg, args...)
}

// write creates and sends a log record to the handler.
// - Adds trace ID if available
// - Captures caller information based on the given depth
func (log *Logger) write(ctx context.Context, level Level, caller int, msg string, args ...any) {
	slogLevel := slog.Level(level)

	// Check if the log level is enabled
	if !log.handler.Enabled(ctx, slogLevel) {
		return
	}

	// Capture the caller's program counter
	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:])

	// Create a new structured log record
	r := slog.NewRecord(time.Now(), slogLevel, msg, pcs[0])

	// Append trace ID if a function is provided
	if log.traceIDFn != nil {
		args = append(args, "trace_id", log.traceIDFn(ctx))
	}

	// Add additional structured attributes
	r.Add(args...)

	// Send the log record to the handler
	log.handler.Handle(ctx, r)
}

// new initializes a Logger with JSON output, optional event hooks, and service tagging.
func new(w io.Writer, minLevel Level, serviceName string, traceIDFn TraceIDFn, events Events) *Logger {
	// ReplaceAttr function to customize source file formatting
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				// Use only the file name and line number
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{
					Key:   "file",
					Value: slog.StringValue(v),
				}
			}
		}
		return a
	}

	// Create a JSON handler with custom options
	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.Level(minLevel),
		ReplaceAttr: f,
	}))

	// Wrap handler with event hooks if provided
	if events.Debug != nil || events.Info != nil || events.Warn != nil || events.Error != nil {
		handler = newLogHandler(handler, events)
	}

	// Add service name as a constant log attribute
	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}
	handler = handler.WithAttrs(attrs)

	return &Logger{
		discard:   w == io.Discard,
		handler:   handler,
		traceIDFn: traceIDFn,
	}
}
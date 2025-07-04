package logger

import (
	"context"
)

// LogLevel represents the severity of a log entry
type LogLevel int

const (
	// LogLevelDebug is for debug information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for informational messages
	LogLevelInfo
	// LogLevelWarn is for warning messages
	LogLevelWarn
	// LogLevelError is for error messages
	LogLevelError
	// LogLevelFatal is for fatal errors
	LogLevelFatal
)

// Logger is the interface for logging in DealDone
type Logger interface {
	// Debug logs a debug message
	Debug(format string, args ...interface{})
	// Info logs an informational message
	Info(format string, args ...interface{})
	// Warn logs a warning message
	Warn(format string, args ...interface{})
	// Error logs an error message
	Error(format string, args ...interface{})
	// Fatal logs a fatal error and exits
	Fatal(format string, args ...interface{})

	// WithContext returns a logger with context
	WithContext(ctx context.Context) Logger
	// WithField returns a logger with a single field
	WithField(key string, value interface{}) Logger
	// WithFields returns a logger with multiple fields
	WithFields(fields map[string]interface{}) Logger
}

// StructuredLogger is an extended logger interface with structured logging support
type StructuredLogger interface {
	Logger

	// DebugContext logs a debug message with context
	DebugContext(ctx context.Context, msg string, fields ...Field)
	// InfoContext logs an info message with context
	InfoContext(ctx context.Context, msg string, fields ...Field)
	// WarnContext logs a warning message with context
	WarnContext(ctx context.Context, msg string, fields ...Field)
	// ErrorContext logs an error message with context
	ErrorContext(ctx context.Context, msg string, fields ...Field)
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a boolean field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

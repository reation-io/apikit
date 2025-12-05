// Package logger provides structured logging for the OpenAPI generator
package logger

import (
	"io"
	"log/slog"
	"os"
)

// Level represents the logging level
type Level = slog.Level

// Log levels
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Logger wraps slog.Logger for structured logging
type Logger struct {
	*slog.Logger
	level Level
}

// globalLogger is the default logger instance
var globalLogger = New(os.Stderr, LevelInfo)

// New creates a new Logger with the specified output and level
func New(w io.Writer, level Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(w, opts)
	return &Logger{
		Logger: slog.New(handler),
		level:  level,
	}
}

// NewJSON creates a new Logger with JSON output
func NewJSON(w io.Writer, level Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(w, opts)
	return &Logger{
		Logger: slog.New(handler),
		level:  level,
	}
}

// SetDefault sets the global default logger
func SetDefault(l *Logger) {
	globalLogger = l
}

// Default returns the global default logger
func Default() *Logger {
	return globalLogger
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// With returns a new Logger with the given attributes
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		level:  l.level,
	}
}

// WithGroup returns a new Logger with the given group name
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: l.Logger.WithGroup(name),
		level:  l.level,
	}
}

// Global logging functions using the default logger

// Debug logs a debug message
func Debug(msg string, args ...any) {
	globalLogger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	globalLogger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	globalLogger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	globalLogger.Error(msg, args...)
}

// NopLogger returns a logger that discards all output
func NopLogger() *Logger {
	return New(io.Discard, LevelError)
}

// Attr creates a structured log attribute
func Attr(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// String creates a string attribute
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

// Int creates an int attribute
func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

// Bool creates a bool attribute
func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}

// Err creates an error attribute
func Err(err error) slog.Attr {
	return slog.Any("error", err)
}


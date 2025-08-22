// Package logger provides structured logging functionality for the VJVector database.
// It uses Go's standard library slog package for consistent logging across the application.
package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

// Init initializes the logger with the specified level
func Init(level string) {
	// Parse log level
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create handler with JSON formatting
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Create logger
	log = slog.New(handler)

	// Set as default logger
	slog.SetDefault(log)
}

// Get returns the logger instance
func Get() *slog.Logger {
	if log == nil {
		Init("info")
	}
	return log
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, args ...any) {
	Get().Error(msg, args...)
	os.Exit(1)
}

// WithField adds a field to the logger
func WithField(key string, value any) *slog.Logger {
	return Get().With(key, value)
}

// WithFields adds multiple fields to the logger
func WithFields(fields map[string]any) *slog.Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return Get().With(args...)
}

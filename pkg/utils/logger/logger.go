package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the logger with the specified level
func Init(level string) {
	log = logrus.New()
	
	// Parse log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)
	
	// Set output to stdout
	log.SetOutput(os.Stdout)
	
	// Set formatter
	log.SetFormatter(&logrus.JSONFormatter{})
}

// Get returns the logger instance
func Get() *logrus.Logger {
	if log == nil {
		Init("info")
	}
	return log
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Get().Debug(args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	Get().Info(args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Get().Warn(args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Get().Error(args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

// WithField adds a field to the logger
func WithField(key string, value interface{}) *logrus.Entry {
	return Get().WithField(key, value)
}

// WithFields adds multiple fields to the logger
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Get().WithFields(fields)
}

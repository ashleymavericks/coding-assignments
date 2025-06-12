package logger

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger interface defines our logging contract
// Go Concept: Interface segregation - define what behavior we need
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	// WithFields creates a new logger with additional fields
	WithFields(fields ...Field) Logger

	// WithError creates a new logger with an error field
	WithError(err error) Logger
}

// Field represents a key-value pair for structured logging
// Go Concept: Simple struct for encapsulating data
type Field struct {
	Key   string
	Value interface{}
}

// logrusLogger implements Logger using logrus
// Go Concept: Struct that implements an interface
type logrusLogger struct {
	logger *logrus.Entry
}

// New creates a new logger instance
// Go Concept: Constructor function pattern
func New(level, format string) Logger {
	// Create base logrus logger
	baseLogger := logrus.New()

	// Set log level
	switch level {
	case "debug":
		baseLogger.SetLevel(logrus.DebugLevel)
	case "info":
		baseLogger.SetLevel(logrus.InfoLevel)
	case "warn":
		baseLogger.SetLevel(logrus.WarnLevel)
	case "error":
		baseLogger.SetLevel(logrus.ErrorLevel)
	default:
		baseLogger.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	if format == "json" {
		baseLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		baseLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Set output to stdout
	baseLogger.SetOutput(os.Stdout)

	return &logrusLogger{
		logger: logrus.NewEntry(baseLogger),
	}
}

// NewWithWriter creates a logger with custom writer
// Go Concept: Function overloading pattern with different constructors
func NewWithWriter(level, format string, writer io.Writer) Logger {
	logger := New(level, format).(*logrusLogger)
	logger.logger.Logger.SetOutput(writer)
	return logger
}

// Debug logs a debug message
// Go Concept: Method implementation for interface
func (l *logrusLogger) Debug(msg string, fields ...Field) {
	l.logWithFields(l.logger.Debug, msg, fields...)
}

// Info logs an info message
func (l *logrusLogger) Info(msg string, fields ...Field) {
	l.logWithFields(l.logger.Info, msg, fields...)
}

// Warn logs a warning message
func (l *logrusLogger) Warn(msg string, fields ...Field) {
	l.logWithFields(l.logger.Warn, msg, fields...)
}

// Error logs an error message
func (l *logrusLogger) Error(msg string, fields ...Field) {
	l.logWithFields(l.logger.Error, msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *logrusLogger) Fatal(msg string, fields ...Field) {
	l.logWithFields(l.logger.Fatal, msg, fields...)
}

// WithFields creates a new logger with additional fields
// Go Concept: Method chaining and immutability pattern
func (l *logrusLogger) WithFields(fields ...Field) Logger {
	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}

	return &logrusLogger{
		logger: l.logger.WithFields(logrusFields),
	}
}

// WithError creates a new logger with an error field
func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{
		logger: l.logger.WithError(err),
	}
}

// logWithFields is a helper method for logging with fields
// Go Concept: Private method for internal use
func (l *logrusLogger) logWithFields(logFunc func(args ...interface{}), msg string, fields ...Field) {
	if len(fields) == 0 {
		logFunc(msg)
		return
	}

	// Convert fields to logrus fields
	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}

	l.logger.WithFields(logrusFields).Log(l.logger.Logger.Level, msg)
}

// Helper functions for creating fields
// Go Concept: Package-level utility functions

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

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Global logger instance for package-level logging
// Go Concept: Package-level variable for singleton pattern
var defaultLogger Logger

// Init initializes the default logger
func Init(level, format string) {
	defaultLogger = New(level, format)
}

// Package-level logging functions that use the default logger
// Go Concept: Convenience functions for common use cases

func Debug(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...Field) {
	if defaultLogger != nil {
		defaultLogger.Fatal(msg, fields...)
	}
}

func WithFields(fields ...Field) Logger {
	if defaultLogger != nil {
		return defaultLogger.WithFields(fields...)
	}
	return New("info", "text")
}

func WithError(err error) Logger {
	if defaultLogger != nil {
		return defaultLogger.WithError(err)
	}
	return New("info", "text")
}

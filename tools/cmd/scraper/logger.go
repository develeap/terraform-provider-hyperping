package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LogLevel represents logging severity
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogFormat represents output format
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// Logger provides structured logging
type Logger struct {
	output io.Writer
	format LogFormat
	level  LogLevel
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new structured logger
func NewLogger(output io.Writer, format LogFormat) *Logger {
	return &Logger{
		output: output,
		format: format,
		level:  LogLevelInfo,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// shouldLog determines if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}
	return levels[level] >= levels[l.level]
}

// log writes a structured log entry
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     string(level),
		Message:   message,
		Fields:    fields,
	}

	if l.format == LogFormatJSON {
		data, err := json.Marshal(entry)
		if err != nil {
			// Fallback to simple output on marshal error
			fmt.Fprintf(l.output, "[%s] %s: %s\n", entry.Timestamp, entry.Level, entry.Message)
			return
		}
		fmt.Fprintln(l.output, string(data))
	} else {
		// Text format with optional fields
		output := fmt.Sprintf("[%s] %s: %s", entry.Timestamp, entry.Level, entry.Message)
		if len(fields) > 0 {
			output += fmt.Sprintf(" %v", fields)
		}
		fmt.Fprintln(l.output, output)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelDebug, message, f)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelInfo, message, f)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelWarn, message, f)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LogLevelError, message, f)
}

// GetLogFormat returns the log format from environment or default to text
func GetLogFormat() LogFormat {
	format := os.Getenv("LOG_FORMAT")
	if format == "json" {
		return LogFormatJSON
	}
	return LogFormatText
}

// GetLogLevel returns the log level from environment or default to info
func GetLogLevel() LogLevel {
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		return LogLevelDebug
	case "warn":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// Global logger instance (will be initialized in main)
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger() {
	format := GetLogFormat()
	globalLogger = NewLogger(os.Stdout, format)
	globalLogger.SetLevel(GetLogLevel())

	// Setup standard log package to use our logger for backward compatibility
	log.SetOutput(os.Stdout)
	log.SetFlags(0) // We handle timestamps in structured logger
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		InitGlobalLogger()
	}
	return globalLogger
}

package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, LogFormatText)

	if logger == nil {
		t.Fatal("Expected logger to be created")
	}
	if logger.format != LogFormatText {
		t.Errorf("Expected format=text, got %s", logger.format)
	}
	if logger.level != LogLevelInfo {
		t.Errorf("Expected level=info, got %s", logger.level)
	}
}

func TestLoggerTextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, LogFormatText)

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "info") {
		t.Errorf("Expected output to contain 'info', got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
}

func TestLoggerJSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, LogFormatJSON)

	logger.Info("test message", map[string]interface{}{
		"key": "value",
	})

	output := buf.String()

	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if entry.Level != "info" {
		t.Errorf("Expected level=info, got %s", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("Expected message='test message', got %s", entry.Message)
	}
	if entry.Fields["key"] != "value" {
		t.Errorf("Expected field key=value, got %v", entry.Fields["key"])
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name         string
		level        LogLevel
		logLevel     LogLevel
		shouldLog    bool
	}{
		{"debug logs at debug level", LogLevelDebug, LogLevelDebug, true},
		{"debug doesn't log at info level", LogLevelDebug, LogLevelInfo, false},
		{"info logs at info level", LogLevelInfo, LogLevelInfo, true},
		{"info logs at debug level", LogLevelInfo, LogLevelDebug, true},
		{"warn logs at warn level", LogLevelWarn, LogLevelWarn, true},
		{"error logs at error level", LogLevelError, LogLevelError, true},
		{"info doesn't log at error level", LogLevelInfo, LogLevelError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(buf, LogFormatText)
			logger.SetLevel(tt.logLevel)

			switch tt.level {
			case LogLevelDebug:
				logger.Debug("test")
			case LogLevelInfo:
				logger.Info("test")
			case LogLevelWarn:
				logger.Warn("test")
			case LogLevelError:
				logger.Error("test")
			}

			output := buf.String()
			logged := len(output) > 0

			if logged != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, got %v (output: %s)", tt.shouldLog, logged, output)
			}
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, LogFormatJSON)

	fields := map[string]interface{}{
		"url":      "https://example.com",
		"duration": 1.5,
		"status":   200,
	}

	logger.Info("request completed", fields)

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry.Fields["url"] != "https://example.com" {
		t.Errorf("Expected url field, got %v", entry.Fields["url"])
	}
	if entry.Fields["duration"] != 1.5 {
		t.Errorf("Expected duration=1.5, got %v", entry.Fields["duration"])
	}
	if entry.Fields["status"] != float64(200) { // JSON numbers are float64
		t.Errorf("Expected status=200, got %v", entry.Fields["status"])
	}
}

func TestGetLogFormat(t *testing.T) {
	// Test default
	format := GetLogFormat()
	if format != LogFormatText {
		t.Errorf("Expected default format=text, got %s", format)
	}

	// Test with env var (would need to set env in real test)
	// Skipping env var test as it requires environment manipulation
}

func TestGetLogLevel(t *testing.T) {
	// Test default
	level := GetLogLevel()
	if level != LogLevelInfo {
		t.Errorf("Expected default level=info, got %s", level)
	}
}

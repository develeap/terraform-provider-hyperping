// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package recovery

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	t.Run("normal mode", func(t *testing.T) {
		logger, err := NewLogger(false)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		if logger.debugMode {
			t.Error("Expected debug mode to be false")
		}

		if logger.logFile != nil {
			t.Error("Expected log file to be nil in normal mode")
		}
	})

	t.Run("debug mode", func(t *testing.T) {
		logger, err := NewLogger(true)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}
		defer logger.Close()

		if !logger.debugMode {
			t.Error("Expected debug mode to be true")
		}

		if logger.logFile == nil {
			t.Error("Expected log file to be created in debug mode")
		}

		if logger.GetLogPath() == "" {
			t.Error("Expected log path to be set in debug mode")
		}

		// Verify log file exists
		if _, err := os.Stat(logger.GetLogPath()); os.IsNotExist(err) {
			t.Errorf("Log file does not exist: %s", logger.GetLogPath())
		}

		// Cleanup
		os.Remove(logger.GetLogPath())
	})
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		writer:    &buf,
		debugMode: true,
	}

	logger.Debug("Debug message: %s", "test")
	output := buf.String()

	if !strings.Contains(output, "DEBUG") {
		t.Error("Expected output to contain DEBUG")
	}
	if !strings.Contains(output, "Debug message: test") {
		t.Error("Expected output to contain the debug message")
	}

	buf.Reset()
	logger.Info("Info message")
	output = buf.String()

	if !strings.Contains(output, "INFO") {
		t.Error("Expected output to contain INFO")
	}

	buf.Reset()
	logger.Warn("Warning message")
	output = buf.String()

	if !strings.Contains(output, "WARN") {
		t.Error("Expected output to contain WARN")
	}

	buf.Reset()
	logger.Error("Error message")
	output = buf.String()

	if !strings.Contains(output, "ERROR") {
		t.Error("Expected output to contain ERROR")
	}
}

func TestLoggerDebugMode(t *testing.T) {
	var buf bytes.Buffer

	logger := &Logger{
		writer:    &buf,
		debugMode: false,
	}

	logger.Debug("This should not appear")
	output := buf.String()

	if output != "" {
		t.Error("Expected no output in non-debug mode")
	}

	// Info should still work
	logger.Info("This should appear")
	output = buf.String()

	if output == "" {
		t.Error("Expected info message to appear")
	}
}

func TestExponentialBackoff(t *testing.T) {
	t.Run("success on first try", func(t *testing.T) {
		backoff := DefaultBackoff()
		ctx := context.Background()

		callCount := 0
		err := backoff.Retry(ctx, func() error {
			callCount++
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if callCount != 1 {
			t.Errorf("Expected 1 call, got %d", callCount)
		}
	})

	t.Run("success after retries", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			MaxRetries:  3,
			InitialWait: 10 * time.Millisecond,
			MaxWait:     100 * time.Millisecond,
			Factor:      2.0,
		}
		ctx := context.Background()

		callCount := 0
		err := backoff.Retry(ctx, func() error {
			callCount++
			if callCount < 3 {
				return errors.New("temporary error")
			}
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if callCount != 3 {
			t.Errorf("Expected 3 calls, got %d", callCount)
		}
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			MaxRetries:  2,
			InitialWait: 10 * time.Millisecond,
			MaxWait:     100 * time.Millisecond,
			Factor:      2.0,
		}
		ctx := context.Background()

		callCount := 0
		err := backoff.Retry(ctx, func() error {
			callCount++
			return errors.New("persistent error")
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if callCount != 3 { // MaxRetries + 1
			t.Errorf("Expected 3 calls, got %d", callCount)
		}

		if !strings.Contains(err.Error(), "max retries exceeded") {
			t.Errorf("Expected 'max retries exceeded' in error, got: %v", err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		backoff := &ExponentialBackoff{
			MaxRetries:  5,
			InitialWait: 100 * time.Millisecond,
			MaxWait:     1 * time.Second,
			Factor:      2.0,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		callCount := 0
		err := backoff.Retry(ctx, func() error {
			callCount++
			return errors.New("error")
		})

		if err == nil {
			t.Error("Expected context error, got nil")
		}

		// Should only call once before context timeout
		if callCount > 2 {
			t.Errorf("Expected at most 2 calls due to context timeout, got %d", callCount)
		}
	})
}

func TestAPIValidator(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		writer:    &buf,
		debugMode: true,
	}

	validator := NewAPIValidator(logger)

	t.Run("successful validation", func(t *testing.T) {
		result := validator.ValidateSourceAPI(context.Background(), "TestAPI", func(ctx context.Context) error {
			return nil
		})

		if !result.Valid {
			t.Error("Expected validation to be valid")
		}
		if !result.CanConnect {
			t.Error("Expected CanConnect to be true")
		}
		if !result.IsAuthenticated {
			t.Error("Expected IsAuthenticated to be true")
		}
	})

	t.Run("failed validation", func(t *testing.T) {
		result := validator.ValidateSourceAPI(context.Background(), "TestAPI", func(ctx context.Context) error {
			return errors.New("connection failed")
		})

		if result.Valid {
			t.Error("Expected validation to be invalid")
		}
		if result.CanConnect {
			t.Error("Expected CanConnect to be false")
		}
		if result.ErrorMessage == "" {
			t.Error("Expected error message to be set")
		}
	})
}

func TestDefaultBackoff(t *testing.T) {
	backoff := DefaultBackoff()

	if backoff.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", backoff.MaxRetries)
	}

	if backoff.InitialWait != 1*time.Second {
		t.Errorf("Expected InitialWait to be 1s, got %v", backoff.InitialWait)
	}

	if backoff.MaxWait != 30*time.Second {
		t.Errorf("Expected MaxWait to be 30s, got %v", backoff.MaxWait)
	}

	if backoff.Factor != 2.0 {
		t.Errorf("Expected Factor to be 2.0, got %f", backoff.Factor)
	}
}

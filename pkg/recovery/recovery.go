// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package recovery provides error recovery utilities for migration tools
package recovery

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Logger provides debug logging capabilities
type Logger struct {
	writer      io.Writer
	debugMode   bool
	logFile     *os.File
	logFilePath string
}

// NewLogger creates a new logger
func NewLogger(debugMode bool) (*Logger, error) {
	logger := &Logger{
		writer:    os.Stderr,
		debugMode: debugMode,
	}

	if debugMode {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		logDir := filepath.Join(homeDir, ".hyperping-migrate", "logs")
		//nolint:govet
		if err := os.MkdirAll(logDir, 0o700); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		timestamp := time.Now().UTC().Format("20060102-150405")
		logFilePath := filepath.Join(logDir, fmt.Sprintf("migration-%s.log", timestamp))

		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return nil, fmt.Errorf("failed to create log file: %w", err)
		}

		logger.logFile = logFile
		logger.logFilePath = logFilePath
		logger.writer = io.MultiWriter(os.Stderr, logFile)
	}

	return logger, nil
}

// Close closes the log file if open
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.debugMode {
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.writer, "[%s] DEBUG: %s\n", timestamp, message)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.writer, "[%s] INFO: %s\n", timestamp, message)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.writer, "[%s] WARN: %s\n", timestamp, message)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.writer, "[%s] ERROR: %s\n", timestamp, message)
}

// GetLogPath returns the path to the log file (empty if not in debug mode)
func (l *Logger) GetLogPath() string {
	return l.logFilePath
}

// APIValidator validates API connectivity and authentication
type APIValidator struct {
	logger *Logger
}

// NewAPIValidator creates a new API validator
func NewAPIValidator(logger *Logger) *APIValidator {
	return &APIValidator{logger: logger}
}

// ValidationResult contains the result of API validation
type ValidationResult struct {
	Valid           bool
	CanConnect      bool
	IsAuthenticated bool
	RateLimitOK     bool
	ErrorMessage    string
	Warnings        []string
}

// ValidateSourceAPI validates connection to source monitoring service
func (v *APIValidator) ValidateSourceAPI(ctx context.Context, serviceName string, validator func(context.Context) error) ValidationResult {
	v.logger.Debug("Validating %s API connectivity...", serviceName)

	result := ValidationResult{
		Valid:      true,
		CanConnect: true,
	}

	if err := validator(ctx); err != nil {
		result.Valid = false
		result.CanConnect = false
		result.ErrorMessage = fmt.Sprintf("Failed to connect to %s API: %v", serviceName, err)
		v.logger.Error("%s", result.ErrorMessage)
		return result
	}

	result.IsAuthenticated = true
	v.logger.Debug("%s API validation successful", serviceName)

	return result
}

// ConfirmAction prompts user for confirmation
func ConfirmAction(prompt string, defaultYes bool) bool {
	var response string
	defaultStr := "y/N"
	if defaultYes {
		defaultStr = "Y/n"
	}

	fmt.Fprintf(os.Stderr, "%s [%s]: ", prompt, defaultStr)
	// Ignore error from Scanln - user input is optional
	fmt.Scanln(&response) //nolint:errcheck

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}

// ExponentialBackoff implements exponential backoff retry logic
type ExponentialBackoff struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Factor      float64
}

// DefaultBackoff returns a default backoff configuration
func DefaultBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		MaxRetries:  3,
		InitialWait: 1 * time.Second,
		MaxWait:     30 * time.Second,
		Factor:      2.0,
	}
}

// Retry executes a function with exponential backoff
func (b *ExponentialBackoff) Retry(ctx context.Context, fn func() error) error {
	var lastErr error
	wait := b.InitialWait

	for attempt := 0; attempt <= b.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}

			wait = time.Duration(float64(wait) * b.Factor)
			if wait > b.MaxWait {
				wait = b.MaxWait
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

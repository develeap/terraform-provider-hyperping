// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package terraflyerrors

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEnhancedError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *EnhancedError
		contains []string
	}{
		{
			name: "basic error",
			err: &EnhancedError{
				Title:       "Test Error",
				Description: "This is a test error",
			},
			contains: []string{
				"❌ Test Error",
				"This is a test error",
			},
		},
		{
			name: "error with context",
			err: &EnhancedError{
				Title:       "Validation Failed",
				Description: "Invalid field value",
				Operation:   "create",
				Resource:    "hyperping_monitor.test",
				Field:       "frequency",
			},
			contains: []string{
				"❌ Validation Failed",
				"Invalid field value",
				"Resource:  hyperping_monitor.test",
				"Operation: create",
				"Field:     frequency",
			},
		},
		{
			name: "error with suggestions",
			err: &EnhancedError{
				Title:       "Rate Limit Exceeded",
				Description: "Too many requests",
				Suggestions: []string{
					"Reduce parallelism",
					"Wait before retrying",
				},
			},
			contains: []string{
				"💡 Suggestions:",
				"Reduce parallelism",
				"Wait before retrying",
			},
		},
		{
			name: "error with commands",
			err: &EnhancedError{
				Title:    "Auth Failed",
				Commands: []string{"echo $HYPERPING_API_KEY", "terraform plan"},
			},
			contains: []string{
				"🔧 Try:",
				"$ echo $HYPERPING_API_KEY",
				"$ terraform plan",
			},
		},
		{
			name: "error with examples",
			err: &EnhancedError{
				Title:    "Invalid Frequency",
				Examples: []string{"frequency = 60", "frequency = 300"},
			},
			contains: []string{
				"📝 Examples:",
				"frequency = 60",
				"frequency = 300",
			},
		},
		{
			name: "error with doc links",
			err: &EnhancedError{
				Title:    "Documentation Needed",
				DocLinks: []string{"https://example.com/docs", "https://example.com/guides"},
			},
			contains: []string{
				"📚 Documentation:",
				"https://example.com/docs",
				"https://example.com/guides",
			},
		},
		{
			name: "error with retry after",
			err: &EnhancedError{
				Title:      "Rate Limited",
				RetryAfter: ptrDuration(30 * time.Second),
			},
			contains: []string{
				"⏰ Auto-retry after: 30s",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.err.Error()

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Error output missing expected string:\nExpected: %s\nGot: %s",
						expected, output)
				}
			}
		})
	}
}

// extractEnhanced is a test helper that unwraps an EnhancedError from an error.
func extractEnhanced(t *testing.T, err error) *EnhancedError {
	t.Helper()
	var enhanced *EnhancedError
	if !errors.As(err, &enhanced) {
		t.Fatal("Expected EnhancedError")
	}
	return enhanced
}

// TestEnhanceError_nilInput verifies that a nil error returns nil.
func TestEnhanceError_nilInput(t *testing.T) {
	result := EnhanceError(nil, CategoryAuth)
	if result != nil {
		t.Errorf("Expected nil error, got: %v", result)
	}
}

// TestEnhanceError_authDefaults verifies auth category defaults are applied.
func TestEnhanceError_authDefaults(t *testing.T) {
	result := EnhanceError(errors.New("unauthorized"), CategoryAuth)
	enhanced := extractEnhanced(t, result)
	if enhanced.Title != "Authentication Failed" {
		t.Errorf("Expected auth title, got: %s", enhanced.Title)
	}
	if len(enhanced.Suggestions) == 0 {
		t.Error("Expected auth suggestions")
	}
	if len(enhanced.DocLinks) == 0 {
		t.Error("Expected auth doc links")
	}
}

// TestEnhanceError_rateLimitDefaults verifies rate limit category defaults.
func TestEnhanceError_rateLimitDefaults(t *testing.T) {
	result := EnhanceError(errors.New("rate limited"), CategoryRateLimit)
	enhanced := extractEnhanced(t, result)
	if enhanced.Title != "Rate Limit Exceeded" {
		t.Errorf("Expected rate limit title, got: %s", enhanced.Title)
	}
	if !enhanced.Retryable {
		t.Error("Expected rate limit to be retryable")
	}
}

// TestEnhanceError_categoryDefaults verifies that each remaining category has the correct title.
func TestEnhanceError_categoryDefaults(t *testing.T) {
	tests := []struct {
		name          string
		category      ErrorCategory
		wantTitle     string
		wantRetryable bool
	}{
		{"validation", CategoryValidation, "Validation Error", false},
		{"not found", CategoryNotFound, "Resource Not Found", false},
		{"server", CategoryServer, "", true},
		{"network", CategoryNetwork, "Network Error", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnhanceError(errors.New("test error"), tt.category)
			enhanced := extractEnhanced(t, result)
			if tt.wantTitle != "" && enhanced.Title != tt.wantTitle {
				t.Errorf("expected title %q, got: %s", tt.wantTitle, enhanced.Title)
			}
			if tt.wantRetryable && !enhanced.Retryable {
				t.Errorf("expected %s error to be retryable", tt.name)
			}
		})
	}
}

// TestEnhanceError_customOptions verifies that options override category defaults.
func TestEnhanceError_customOptions(t *testing.T) {
	result := EnhanceError(
		errors.New("test"),
		CategoryAuth,
		WithTitle("Custom Title"),
		WithDescription("Custom description"),
		WithOperation("create"),
		WithResource("hyperping_monitor.test"),
	)
	enhanced := extractEnhanced(t, result)
	if enhanced.Title != "Custom Title" {
		t.Errorf("Expected custom title, got: %s", enhanced.Title)
	}
	if enhanced.Operation != "create" {
		t.Errorf("Expected create operation, got: %s", enhanced.Operation)
	}
}

func TestEnhancedError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	enhanced := &EnhancedError{
		Title:      "Enhanced",
		Underlying: underlying,
	}

	unwrapped := errors.Unwrap(enhanced)
	if unwrapped != underlying {
		t.Errorf("Expected unwrapped error to be underlying, got: %v", unwrapped)
	}
}

func TestEnhancementOptions(t *testing.T) {
	tests := []struct {
		name     string
		opt      EnhancementOption
		validate func(*testing.T, *EnhancedError)
	}{
		{
			name: "WithTitle",
			opt:  WithTitle("Test Title"),
			validate: func(t *testing.T, e *EnhancedError) {
				if e.Title != "Test Title" {
					t.Errorf("Expected title 'Test Title', got: %s", e.Title)
				}
			},
		},
		{
			name: "WithDescription",
			opt:  WithDescription("Test description"),
			validate: func(t *testing.T, e *EnhancedError) {
				if e.Description != "Test description" {
					t.Errorf("Expected description, got: %s", e.Description)
				}
			},
		},
		{
			name: "WithOperation",
			opt:  WithOperation("create"),
			validate: func(t *testing.T, e *EnhancedError) {
				if e.Operation != "create" {
					t.Errorf("Expected operation 'create', got: %s", e.Operation)
				}
			},
		},
		{
			name: "WithResource",
			opt:  WithResource("hyperping_monitor.test"),
			validate: func(t *testing.T, e *EnhancedError) {
				if e.Resource != "hyperping_monitor.test" {
					t.Errorf("Expected resource, got: %s", e.Resource)
				}
			},
		},
		{
			name: "WithField",
			opt:  WithField("frequency"),
			validate: func(t *testing.T, e *EnhancedError) {
				if e.Field != "frequency" {
					t.Errorf("Expected field 'frequency', got: %s", e.Field)
				}
			},
		},
		{
			name: "WithSuggestions",
			opt:  WithSuggestions("suggestion 1", "suggestion 2"),
			validate: func(t *testing.T, e *EnhancedError) {
				if len(e.Suggestions) != 2 {
					t.Errorf("Expected 2 suggestions, got: %d", len(e.Suggestions))
				}
			},
		},
		{
			name: "WithCommands",
			opt:  WithCommands("command 1", "command 2"),
			validate: func(t *testing.T, e *EnhancedError) {
				if len(e.Commands) != 2 {
					t.Errorf("Expected 2 commands, got: %d", len(e.Commands))
				}
			},
		},
		{
			name: "WithDocLinks",
			opt:  WithDocLinks("https://example.com"),
			validate: func(t *testing.T, e *EnhancedError) {
				if len(e.DocLinks) != 1 {
					t.Errorf("Expected 1 doc link, got: %d", len(e.DocLinks))
				}
			},
		},
		{
			name: "WithExamples",
			opt:  WithExamples("example 1"),
			validate: func(t *testing.T, e *EnhancedError) {
				if len(e.Examples) != 1 {
					t.Errorf("Expected 1 example, got: %d", len(e.Examples))
				}
			},
		},
		{
			name: "WithRetryable",
			opt:  WithRetryable(true),
			validate: func(t *testing.T, e *EnhancedError) {
				if !e.Retryable {
					t.Error("Expected retryable to be true")
				}
			},
		},
		{
			name: "WithRetryAfter",
			opt:  WithRetryAfter(30 * time.Second),
			validate: func(t *testing.T, e *EnhancedError) {
				if e.RetryAfter == nil {
					t.Fatal("Expected RetryAfter to be set")
				}
				if *e.RetryAfter != 30*time.Second {
					t.Errorf("Expected 30s, got: %v", *e.RetryAfter)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced := &EnhancedError{}
			tt.opt(enhanced)
			tt.validate(t, enhanced)
		})
	}
}

func TestFormatContext(t *testing.T) {
	tests := []struct {
		name     string
		err      *EnhancedError
		contains []string
		absent   []string
	}{
		{
			name:   "no context fields",
			err:    &EnhancedError{Title: "Test"},
			absent: []string{"Resource:", "Operation:", "Field:"},
		},
		{
			name:     "resource only",
			err:      &EnhancedError{Resource: "hyperping_monitor.prod"},
			contains: []string{"Resource:  hyperping_monitor.prod"},
			absent:   []string{"Operation:", "Field:"},
		},
		{
			name:     "operation only",
			err:      &EnhancedError{Operation: "create"},
			contains: []string{"Operation: create"},
			absent:   []string{"Resource:", "Field:"},
		},
		{
			name:     "field only",
			err:      &EnhancedError{Field: "frequency"},
			contains: []string{"Field:     frequency"},
			absent:   []string{"Resource:", "Operation:"},
		},
		{
			name: "all context fields",
			err: &EnhancedError{
				Resource:  "hyperping_monitor.prod",
				Operation: "update",
				Field:     "url",
			},
			contains: []string{
				"Resource:  hyperping_monitor.prod",
				"Operation: update",
				"Field:     url",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			tt.err.formatContext(&b)
			output := b.String()

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("formatContext output missing %q; got: %q", want, output)
				}
			}
			for _, notWant := range tt.absent {
				if strings.Contains(output, notWant) {
					t.Errorf("formatContext output unexpectedly contains %q; got: %q", notWant, output)
				}
			}
		})
	}
}

func TestFormatSuggestions(t *testing.T) {
	tests := []struct {
		name     string
		err      *EnhancedError
		contains []string
		absent   []string
	}{
		{
			name:   "no suggestions sections",
			err:    &EnhancedError{Title: "Test"},
			absent: []string{"Suggestions:", "Try:", "Examples:", "Documentation:", "Auto-retry"},
		},
		{
			name:     "suggestions only",
			err:      &EnhancedError{Suggestions: []string{"Fix A", "Fix B"}},
			contains: []string{"💡 Suggestions:", "  • Fix A", "  • Fix B"},
			absent:   []string{"Try:", "Examples:", "Documentation:"},
		},
		{
			name:     "commands only",
			err:      &EnhancedError{Commands: []string{"terraform plan"}},
			contains: []string{"🔧 Try:", "  $ terraform plan"},
			absent:   []string{"Suggestions:", "Examples:"},
		},
		{
			name:     "examples only",
			err:      &EnhancedError{Examples: []string{"frequency = 60"}},
			contains: []string{"📝 Examples:", "  frequency = 60"},
			absent:   []string{"Suggestions:", "Documentation:"},
		},
		{
			name:     "doc links only",
			err:      &EnhancedError{DocLinks: []string{"https://docs.example.com"}},
			contains: []string{"📚 Documentation:", "  https://docs.example.com"},
			absent:   []string{"Suggestions:", "Try:", "Examples:"},
		},
		{
			name:     "retry after only",
			err:      &EnhancedError{RetryAfter: ptrDuration(30 * time.Second)},
			contains: []string{"⏰ Auto-retry after: 30s"},
			absent:   []string{"Suggestions:", "Try:", "Examples:", "Documentation:"},
		},
		{
			name: "all sections present",
			err: &EnhancedError{
				Suggestions: []string{"Suggestion one"},
				Commands:    []string{"cmd --help"},
				Examples:    []string{"example = true"},
				DocLinks:    []string{"https://example.com"},
				RetryAfter:  ptrDuration(60 * time.Second),
			},
			contains: []string{
				"💡 Suggestions:", "  • Suggestion one",
				"🔧 Try:", "  $ cmd --help",
				"📝 Examples:", "  example = true",
				"📚 Documentation:", "  https://example.com",
				"⏰ Auto-retry after: 1m0s",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			tt.err.formatSuggestions(&b)
			output := b.String()

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("formatSuggestions output missing %q; got: %q", want, output)
				}
			}
			for _, notWant := range tt.absent {
				if strings.Contains(output, notWant) {
					t.Errorf("formatSuggestions output unexpectedly contains %q; got: %q", notWant, output)
				}
			}
		})
	}
}

func ptrDuration(d time.Duration) *time.Duration {
	return &d
}

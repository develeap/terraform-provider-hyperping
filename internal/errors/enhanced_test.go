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

func TestEnhanceError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		category ErrorCategory
		opts     []EnhancementOption
		validate func(*testing.T, error)
	}{
		{
			name:     "nil error returns nil",
			err:      nil,
			category: CategoryAuth,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("Expected nil error, got: %v", err)
				}
			},
		},
		{
			name:     "auth error applies defaults",
			err:      errors.New("unauthorized"),
			category: CategoryAuth,
			validate: func(t *testing.T, err error) {
				var enhanced *EnhancedError
				if !errors.As(err, &enhanced) {
					t.Fatal("Expected EnhancedError")
				}
				if enhanced.Title != "Authentication Failed" {
					t.Errorf("Expected auth title, got: %s", enhanced.Title)
				}
				if len(enhanced.Suggestions) == 0 {
					t.Error("Expected auth suggestions")
				}
				if len(enhanced.DocLinks) == 0 {
					t.Error("Expected auth doc links")
				}
			},
		},
		{
			name:     "rate limit error applies defaults",
			err:      errors.New("rate limited"),
			category: CategoryRateLimit,
			validate: func(t *testing.T, err error) {
				var enhanced *EnhancedError
				if !errors.As(err, &enhanced) {
					t.Fatal("Expected EnhancedError")
				}
				if enhanced.Title != "Rate Limit Exceeded" {
					t.Errorf("Expected rate limit title, got: %s", enhanced.Title)
				}
				if !enhanced.Retryable {
					t.Error("Expected rate limit to be retryable")
				}
			},
		},
		{
			name:     "validation error applies defaults",
			err:      errors.New("invalid field"),
			category: CategoryValidation,
			validate: func(t *testing.T, err error) {
				var enhanced *EnhancedError
				if !errors.As(err, &enhanced) {
					t.Fatal("Expected EnhancedError")
				}
				if enhanced.Title != "Validation Error" {
					t.Errorf("Expected validation title, got: %s", enhanced.Title)
				}
			},
		},
		{
			name:     "not found error applies defaults",
			err:      errors.New("not found"),
			category: CategoryNotFound,
			validate: func(t *testing.T, err error) {
				var enhanced *EnhancedError
				if !errors.As(err, &enhanced) {
					t.Fatal("Expected EnhancedError")
				}
				if enhanced.Title != "Resource Not Found" {
					t.Errorf("Expected not found title, got: %s", enhanced.Title)
				}
			},
		},
		{
			name:     "server error applies defaults",
			err:      errors.New("server error"),
			category: CategoryServer,
			validate: func(t *testing.T, err error) {
				var enhanced *EnhancedError
				if !errors.As(err, &enhanced) {
					t.Fatal("Expected EnhancedError")
				}
				if !enhanced.Retryable {
					t.Error("Expected server error to be retryable")
				}
			},
		},
		{
			name:     "network error applies defaults",
			err:      errors.New("network error"),
			category: CategoryNetwork,
			validate: func(t *testing.T, err error) {
				var enhanced *EnhancedError
				if !errors.As(err, &enhanced) {
					t.Fatal("Expected EnhancedError")
				}
				if enhanced.Title != "Network Error" {
					t.Errorf("Expected network title, got: %s", enhanced.Title)
				}
			},
		},
		{
			name:     "custom options override defaults",
			err:      errors.New("test"),
			category: CategoryAuth,
			opts: []EnhancementOption{
				WithTitle("Custom Title"),
				WithDescription("Custom description"),
				WithOperation("create"),
				WithResource("hyperping_monitor.test"),
			},
			validate: func(t *testing.T, err error) {
				var enhanced *EnhancedError
				if !errors.As(err, &enhanced) {
					t.Fatal("Expected EnhancedError")
				}
				if enhanced.Title != "Custom Title" {
					t.Errorf("Expected custom title, got: %s", enhanced.Title)
				}
				if enhanced.Operation != "create" {
					t.Errorf("Expected create operation, got: %s", enhanced.Operation)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enhanced := EnhanceError(tt.err, tt.category, tt.opts...)
			tt.validate(t, enhanced)
		})
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

func ptrDuration(d time.Duration) *time.Duration {
	return &d
}

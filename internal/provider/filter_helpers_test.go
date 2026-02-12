// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMatchesNameRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		pattern  types.String
		expected bool
		wantErr  bool
	}{
		{
			name:     "null pattern matches all",
			input:    "anything",
			pattern:  types.StringNull(),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "unknown pattern matches all",
			input:    "anything",
			pattern:  types.StringUnknown(),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "exact match",
			input:    "api-monitor",
			pattern:  types.StringValue("api-monitor"),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "exact no match",
			input:    "api-monitor",
			pattern:  types.StringValue("db-monitor"),
			expected: false,
			wantErr:  false,
		},
		{
			name:     "regex prefix match",
			input:    "[PROD]-API-Monitor",
			pattern:  types.StringValue("\\[PROD\\]-.*"),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "regex prefix no match",
			input:    "[DEV]-API-Monitor",
			pattern:  types.StringValue("\\[PROD\\]-.*"),
			expected: false,
			wantErr:  false,
		},
		{
			name:     "regex contains match",
			input:    "production-api-service",
			pattern:  types.StringValue(".*api.*"),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "regex case sensitive",
			input:    "Production-API",
			pattern:  types.StringValue("production.*"),
			expected: false,
			wantErr:  false,
		},
		{
			name:     "regex anchored start",
			input:    "test-monitor",
			pattern:  types.StringValue("^test-.*"),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "regex anchored end",
			input:    "monitor-test",
			pattern:  types.StringValue(".*-test$"),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "empty string matches empty pattern",
			input:    "",
			pattern:  types.StringValue(""),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "invalid regex returns error",
			input:    "anything",
			pattern:  types.StringValue("[invalid("),
			expected: false,
			wantErr:  true,
		},
		{
			name:     "invalid regex unclosed bracket",
			input:    "anything",
			pattern:  types.StringValue("[abc"),
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchesNameRegex(tt.input, tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchesNameRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("MatchesNameRegex() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchesExact(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		filter   types.String
		expected bool
	}{
		{
			name:     "null filter matches all",
			value:    "anything",
			filter:   types.StringNull(),
			expected: true,
		},
		{
			name:     "unknown filter matches all",
			value:    "anything",
			filter:   types.StringUnknown(),
			expected: true,
		},
		{
			name:     "exact match",
			value:    "http",
			filter:   types.StringValue("http"),
			expected: true,
		},
		{
			name:     "no match",
			value:    "https",
			filter:   types.StringValue("http"),
			expected: false,
		},
		{
			name:     "case sensitive",
			value:    "HTTP",
			filter:   types.StringValue("http"),
			expected: false,
		},
		{
			name:     "empty matches empty",
			value:    "",
			filter:   types.StringValue(""),
			expected: true,
		},
		{
			name:     "empty value no match",
			value:    "",
			filter:   types.StringValue("http"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesExact(tt.value, tt.filter)
			if got != tt.expected {
				t.Errorf("MatchesExact() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchesExactCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		filter   types.String
		expected bool
	}{
		{
			name:     "null filter matches all",
			value:    "anything",
			filter:   types.StringNull(),
			expected: true,
		},
		{
			name:     "exact match same case",
			value:    "http",
			filter:   types.StringValue("http"),
			expected: true,
		},
		{
			name:     "exact match different case",
			value:    "HTTP",
			filter:   types.StringValue("http"),
			expected: true,
		},
		{
			name:     "mixed case match",
			value:    "HtTp",
			filter:   types.StringValue("http"),
			expected: true,
		},
		{
			name:     "no match",
			value:    "https",
			filter:   types.StringValue("http"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesExactCaseInsensitive(tt.value, tt.filter)
			if got != tt.expected {
				t.Errorf("MatchesExactCaseInsensitive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchesBool(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		filter   types.Bool
		expected bool
	}{
		{
			name:     "null filter matches all",
			value:    true,
			filter:   types.BoolNull(),
			expected: true,
		},
		{
			name:     "unknown filter matches all",
			value:    false,
			filter:   types.BoolUnknown(),
			expected: true,
		},
		{
			name:     "true matches true",
			value:    true,
			filter:   types.BoolValue(true),
			expected: true,
		},
		{
			name:     "false matches false",
			value:    false,
			filter:   types.BoolValue(false),
			expected: true,
		},
		{
			name:     "true does not match false",
			value:    true,
			filter:   types.BoolValue(false),
			expected: false,
		},
		{
			name:     "false does not match true",
			value:    false,
			filter:   types.BoolValue(true),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesBool(tt.value, tt.filter)
			if got != tt.expected {
				t.Errorf("MatchesBool() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchesStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		filter   types.String
		expected bool
	}{
		{
			name:     "null filter matches all",
			values:   []string{"a", "b", "c"},
			filter:   types.StringNull(),
			expected: true,
		},
		{
			name:     "unknown filter matches all",
			values:   []string{"a", "b", "c"},
			filter:   types.StringUnknown(),
			expected: true,
		},
		{
			name:     "empty slice does not match",
			values:   []string{},
			filter:   types.StringValue("a"),
			expected: false,
		},
		{
			name:     "nil slice does not match",
			values:   nil,
			filter:   types.StringValue("a"),
			expected: false,
		},
		{
			name:     "matches first element",
			values:   []string{"http", "https", "tcp"},
			filter:   types.StringValue("http"),
			expected: true,
		},
		{
			name:     "matches middle element",
			values:   []string{"http", "https", "tcp"},
			filter:   types.StringValue("https"),
			expected: true,
		},
		{
			name:     "matches last element",
			values:   []string{"http", "https", "tcp"},
			filter:   types.StringValue("tcp"),
			expected: true,
		},
		{
			name:     "no match",
			values:   []string{"http", "https", "tcp"},
			filter:   types.StringValue("udp"),
			expected: false,
		},
		{
			name:     "case sensitive",
			values:   []string{"http", "https", "tcp"},
			filter:   types.StringValue("HTTP"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesStringSlice(tt.values, tt.filter)
			if got != tt.expected {
				t.Errorf("MatchesStringSlice() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchesStringSliceCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		filter   types.String
		expected bool
	}{
		{
			name:     "null filter matches all",
			values:   []string{"a", "b", "c"},
			filter:   types.StringNull(),
			expected: true,
		},
		{
			name:     "matches different case",
			values:   []string{"http", "https", "tcp"},
			filter:   types.StringValue("HTTP"),
			expected: true,
		},
		{
			name:     "matches mixed case",
			values:   []string{"HtTp", "https", "tcp"},
			filter:   types.StringValue("http"),
			expected: true,
		},
		{
			name:     "no match",
			values:   []string{"http", "https", "tcp"},
			filter:   types.StringValue("udp"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesStringSliceCaseInsensitive(tt.values, tt.filter)
			if got != tt.expected {
				t.Errorf("MatchesStringSliceCaseInsensitive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchesInt64(t *testing.T) {
	tests := []struct {
		name     string
		value    int64
		filter   types.Int64
		expected bool
	}{
		{
			name:     "null filter matches all",
			value:    42,
			filter:   types.Int64Null(),
			expected: true,
		},
		{
			name:     "unknown filter matches all",
			value:    42,
			filter:   types.Int64Unknown(),
			expected: true,
		},
		{
			name:     "exact match",
			value:    100,
			filter:   types.Int64Value(100),
			expected: true,
		},
		{
			name:     "no match",
			value:    100,
			filter:   types.Int64Value(200),
			expected: false,
		},
		{
			name:     "zero matches zero",
			value:    0,
			filter:   types.Int64Value(0),
			expected: true,
		},
		{
			name:     "negative value",
			value:    -10,
			filter:   types.Int64Value(-10),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesInt64(tt.value, tt.filter)
			if got != tt.expected {
				t.Errorf("MatchesInt64() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchesInt64Range(t *testing.T) {
	tests := []struct {
		name     string
		value    int64
		min      types.Int64
		max      types.Int64
		expected bool
	}{
		{
			name:     "no bounds matches all",
			value:    100,
			min:      types.Int64Null(),
			max:      types.Int64Null(),
			expected: true,
		},
		{
			name:     "within range inclusive bounds",
			value:    100,
			min:      types.Int64Value(50),
			max:      types.Int64Value(150),
			expected: true,
		},
		{
			name:     "equals min bound",
			value:    50,
			min:      types.Int64Value(50),
			max:      types.Int64Value(150),
			expected: true,
		},
		{
			name:     "equals max bound",
			value:    150,
			min:      types.Int64Value(50),
			max:      types.Int64Value(150),
			expected: true,
		},
		{
			name:     "below min bound",
			value:    40,
			min:      types.Int64Value(50),
			max:      types.Int64Value(150),
			expected: false,
		},
		{
			name:     "above max bound",
			value:    160,
			min:      types.Int64Value(50),
			max:      types.Int64Value(150),
			expected: false,
		},
		{
			name:     "only min bound set",
			value:    100,
			min:      types.Int64Value(50),
			max:      types.Int64Null(),
			expected: true,
		},
		{
			name:     "only max bound set",
			value:    100,
			min:      types.Int64Null(),
			max:      types.Int64Value(150),
			expected: true,
		},
		{
			name:     "below min when only min set",
			value:    40,
			min:      types.Int64Value(50),
			max:      types.Int64Null(),
			expected: false,
		},
		{
			name:     "above max when only max set",
			value:    160,
			min:      types.Int64Null(),
			max:      types.Int64Value(150),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesInt64Range(tt.value, tt.min, tt.max)
			if got != tt.expected {
				t.Errorf("MatchesInt64Range() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContainsSubstring(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		filter   types.String
		expected bool
	}{
		{
			name:     "null filter matches all",
			value:    "anything",
			filter:   types.StringNull(),
			expected: true,
		},
		{
			name:     "contains substring",
			value:    "production-api-service",
			filter:   types.StringValue("api"),
			expected: true,
		},
		{
			name:     "contains substring different case",
			value:    "production-API-service",
			filter:   types.StringValue("api"),
			expected: true,
		},
		{
			name:     "contains at start",
			value:    "api-service",
			filter:   types.StringValue("api"),
			expected: true,
		},
		{
			name:     "contains at end",
			value:    "service-api",
			filter:   types.StringValue("api"),
			expected: true,
		},
		{
			name:     "does not contain",
			value:    "database-service",
			filter:   types.StringValue("api"),
			expected: false,
		},
		{
			name:     "empty string contains empty",
			value:    "",
			filter:   types.StringValue(""),
			expected: true,
		},
		{
			name:     "all strings contain empty",
			value:    "anything",
			filter:   types.StringValue(""),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsSubstring(tt.value, tt.filter)
			if got != tt.expected {
				t.Errorf("ContainsSubstring() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestApplyAllFilters(t *testing.T) {
	tests := []struct {
		name     string
		filters  []func() bool
		expected bool
	}{
		{
			name:     "empty filter list returns true",
			filters:  []func() bool{},
			expected: true,
		},
		{
			name: "single true filter",
			filters: []func() bool{
				func() bool { return true },
			},
			expected: true,
		},
		{
			name: "single false filter",
			filters: []func() bool{
				func() bool { return false },
			},
			expected: false,
		},
		{
			name: "all true filters",
			filters: []func() bool{
				func() bool { return true },
				func() bool { return true },
				func() bool { return true },
			},
			expected: true,
		},
		{
			name: "first filter false",
			filters: []func() bool{
				func() bool { return false },
				func() bool { return true },
				func() bool { return true },
			},
			expected: false,
		},
		{
			name: "middle filter false",
			filters: []func() bool{
				func() bool { return true },
				func() bool { return false },
				func() bool { return true },
			},
			expected: false,
		},
		{
			name: "last filter false",
			filters: []func() bool{
				func() bool { return true },
				func() bool { return true },
				func() bool { return false },
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyAllFilters(tt.filters...)
			if got != tt.expected {
				t.Errorf("ApplyAllFilters() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Performance test with large dataset
func TestMatchesNameRegexPerformance(t *testing.T) {
	pattern := types.StringValue("\\[PROD\\]-.*-api-.*")

	testCases := []string{
		"[PROD]-service-api-gateway",
		"[DEV]-service-api-gateway",
		"[PROD]-database-mysql",
		"[PROD]-cache-redis-api",
	}

	// Simulate filtering a large dataset
	for i := 0; i < 1000; i++ {
		for _, name := range testCases {
			_, err := MatchesNameRegex(name, pattern)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
	}
}

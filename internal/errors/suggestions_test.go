// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package terraflyerrors

import (
	"strings"
	"testing"
)

func TestValidationSuggestion(t *testing.T) {
	tests := []struct {
		name          string
		field         string
		currentValue  string
		allowedValues []interface{}
		wantContains  []string
	}{
		{
			name:          "basic validation",
			field:         "status",
			currentValue:  "invalid",
			allowedValues: []interface{}{"active", "paused", "deleted"},
			wantContains: []string{
				"Invalid Field Value",
				"status",
				"invalid",
				"active, paused, deleted",
			},
		},
		{
			name:          "finds closest match",
			field:         "severity",
			currentValue:  "critcal",
			allowedValues: []interface{}{"minor", "major", "critical"},
			wantContains: []string{
				"Did you mean 'critical'?",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationSuggestion(tt.field, tt.currentValue, tt.allowedValues)
			output := err.Error()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestFrequencySuggestion(t *testing.T) {
	tests := []struct {
		name         string
		currentValue int64
		wantContains []string
	}{
		{
			name:         "finds closest frequencies",
			currentValue: 45,
			wantContains: []string{
				"Invalid Monitor Frequency",
				"45 seconds (invalid)",
				"30 seconds",
				"60 seconds",
			},
		},
		{
			name:         "invalid zero",
			currentValue: 0,
			wantContains: []string{
				"Invalid Monitor Frequency",
				"10 seconds",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FrequencySuggestion(tt.currentValue)
			output := err.Error()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}

			// Verify examples are present
			if !strings.Contains(output, "Examples:") {
				t.Error("Expected examples section")
			}

			// Verify doc links are present
			if !strings.Contains(output, "Documentation:") {
				t.Error("Expected documentation links")
			}
		})
	}
}

func TestRegionSuggestion(t *testing.T) {
	tests := []struct {
		name         string
		invalidValue string
		wantMatch    string
	}{
		{
			name:         "typo in region",
			invalidValue: "frenkfurt",
			wantMatch:    "frankfurt",
		},
		{
			name:         "partial region name",
			invalidValue: "sing",
			wantMatch:    "singapore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegionSuggestion(tt.invalidValue)
			output := err.Error()

			if !strings.Contains(output, "Invalid Region") {
				t.Error("Expected 'Invalid Region' in title")
			}

			if tt.wantMatch != "" && !strings.Contains(output, tt.wantMatch) {
				t.Errorf("Expected suggestion for %q, got: %s", tt.wantMatch, output)
			}
		})
	}
}

func TestHTTPMethodSuggestion(t *testing.T) {
	tests := []struct {
		name         string
		invalidValue string
		wantContains []string
	}{
		{
			name:         "lowercase method",
			invalidValue: "get",
			wantContains: []string{"Invalid HTTP Method", "GET"},
		},
		{
			name:         "typo in method",
			invalidValue: "PSOT",
			wantContains: []string{"Invalid HTTP Method", "POST"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HTTPMethodSuggestion(tt.invalidValue)
			output := err.Error()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestStatusCodeSuggestion(t *testing.T) {
	tests := []struct {
		name         string
		currentValue string
		wantContains []string
	}{
		{
			name:         "invalid format",
			currentValue: "999",
			wantContains: []string{
				"Invalid Status Code",
				"999",
				`"200"`,
				`"2xx"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StatusCodeSuggestion(tt.currentValue)
			output := err.Error()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestIncidentStatusSuggestion(t *testing.T) {
	tests := []struct {
		name         string
		invalidValue string
		wantMatch    string
	}{
		{
			name:         "typo in status",
			invalidValue: "investgating",
			wantMatch:    "investigating",
		},
		{
			name:         "wrong status",
			invalidValue: "closed",
			wantMatch:    "", // No close match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IncidentStatusSuggestion(tt.invalidValue)
			output := err.Error()

			if !strings.Contains(output, "Invalid Incident Status") {
				t.Error("Expected 'Invalid Incident Status' in title")
			}

			if !strings.Contains(output, "Status workflow") {
				t.Error("Expected status workflow in description")
			}
		})
	}
}

func TestSeveritySuggestion(t *testing.T) {
	tests := []struct {
		name         string
		invalidValue string
		wantMatch    string
	}{
		{
			name:         "typo in severity",
			invalidValue: "criticl",
			wantMatch:    "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SeveritySuggestion(tt.invalidValue)
			output := err.Error()

			if !strings.Contains(output, "Invalid Incident Severity") {
				t.Error("Expected 'Invalid Incident Severity' in title")
			}

			if tt.wantMatch != "" && !strings.Contains(output, tt.wantMatch) {
				t.Errorf("Expected suggestion for %q, got: %s", tt.wantMatch, output)
			}
		})
	}
}

func TestFindClosestString(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		candidates []string
		want       string
	}{
		{
			name:       "exact match",
			target:     "test",
			candidates: []string{"test", "best", "rest"},
			want:       "test",
		},
		{
			name:       "close match",
			target:     "frenkfurt",
			candidates: []string{"london", "frankfurt", "paris"},
			want:       "frankfurt",
		},
		{
			name:       "prefix match",
			target:     "sing",
			candidates: []string{"singapore", "sydney", "tokyo"},
			want:       "singapore",
		},
		{
			name:       "no close match",
			target:     "xyz",
			candidates: []string{"abc", "def", "ghi"},
			want:       "",
		},
		{
			name:       "empty candidates",
			target:     "test",
			candidates: []string{},
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findClosestString(tt.target, tt.candidates)
			if got != tt.want {
				t.Errorf("findClosestString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindClosestFrequencies(t *testing.T) {
	tests := []struct {
		name       string
		target     int64
		candidates []int64
		n          int
		want       []int64
	}{
		{
			name:       "finds closest two",
			target:     45,
			candidates: []int64{10, 20, 30, 60, 120, 300},
			n:          2,
			want:       []int64{30, 60},
		},
		{
			name:       "exact match included",
			target:     60,
			candidates: []int64{10, 30, 60, 120, 300},
			n:          2,
			want:       []int64{60, 30}, // 60 is closest (distance 0)
		},
		{
			name:       "n exceeds candidates",
			target:     100,
			candidates: []int64{10, 60},
			n:          5,
			want:       []int64{60, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findClosestFrequencies(tt.target, tt.candidates, tt.n)
			if len(got) != len(tt.want) {
				t.Errorf("Expected %d results, got %d", len(tt.want), len(got))
			}
			// Check first result (closest)
			if len(got) > 0 && len(tt.want) > 0 && got[0] != tt.want[0] {
				t.Errorf("Expected closest %d, got %d", tt.want[0], got[0])
			}
		})
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		s1   string
		s2   string
		want int
	}{
		{"", "", 0},
		{"test", "", 4},
		{"", "test", 4},
		{"test", "test", 0},
		{"test", "best", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"-"+tt.s2, func(t *testing.T) {
			got := levenshtein(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}

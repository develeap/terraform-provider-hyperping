// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapRegions(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "simple regions",
			input:    []string{"us", "eu"},
			expected: []string{"virginia", "london"},
		},
		{
			name:     "AWS-style regions",
			input:    []string{"us-east-1", "eu-west-1", "ap-southeast-1"},
			expected: []string{"virginia", "london", "singapore"},
		},
		{
			name:     "duplicates removed",
			input:    []string{"us", "us-east", "us-east-1"},
			expected: []string{"virginia"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "mixed case",
			input:    []string{"US", "EU", "Asia"},
			expected: []string{"virginia", "london", "singapore"},
		},
		{
			name:     "unknown regions ignored",
			input:    []string{"unknown", "us", "invalid"},
			expected: []string{"virginia"},
		},
		{
			name:     "all regions",
			input:    []string{"us", "us-west", "eu", "eu-central", "asia", "ap-northeast", "au", "sa", "me"},
			expected: []string{"virginia", "oregon", "london", "frankfurt", "singapore", "tokyo", "sydney", "saopaulo", "bahrain"},
		},
		{
			name:     "middle east aliases",
			input:    []string{"me", "me-south", "me-south-1"},
			expected: []string{"bahrain"},
		},
		{
			name:     "whitespace trimmed",
			input:    []string{" us ", "  eu  "},
			expected: []string{"virginia", "london"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapRegions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultRegions(t *testing.T) {
	regions := DefaultRegions()
	assert.Equal(t, []string{"london", "virginia", "singapore"}, regions)
}

// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapFrequency(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		// Exact matches
		{"exact 10", 10, 10},
		{"exact 20", 20, 20},
		{"exact 30", 30, 30},
		{"exact 60", 60, 60},
		{"exact 120", 120, 120},
		{"exact 180", 180, 180},
		{"exact 300", 300, 300},
		{"exact 600", 600, 600},
		{"exact 1800", 1800, 1800},
		{"exact 3600", 3600, 3600},
		{"exact 21600", 21600, 21600},
		{"exact 43200", 43200, 43200},
		{"exact 86400", 86400, 86400},

		// Rounding cases
		{"5 rounds to 10", 5, 10},
		{"15 rounds to 10", 15, 10},
		{"25 rounds to 20", 25, 20},
		{"45 rounds to 30", 45, 30},
		{"90 rounds to 60", 90, 60},
		{"150 rounds to 120", 150, 120},
		{"350 rounds to 300", 350, 300},
		{"500 rounds to 600", 500, 600},
		{"1000 rounds to 600", 1000, 600},
		{"2700 rounds to 1800", 2700, 1800},

		// Edge cases
		{"0 rounds to 10", 0, 10},
		{"1 rounds to 10", 1, 10},
		{"very large rounds to 86400", 100000, 86400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapFrequency(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAbs(t *testing.T) {
	assert.Equal(t, 5, abs(5))
	assert.Equal(t, 5, abs(-5))
	assert.Equal(t, 0, abs(0))
}

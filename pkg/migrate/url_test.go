// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureURLScheme(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"already https", "https://example.com", "https://example.com"},
		{"already http", "http://example.com", "http://example.com"},
		{"bare domain", "example.com", "https://example.com"},
		{"IP address", "192.168.1.1", "https://192.168.1.1"},
		{"IP with port", "192.168.1.1:8080", "https://192.168.1.1:8080"},
		{"domain with path", "example.com/health", "https://example.com/health"},
		{"empty string", "", "https://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnsureURLScheme(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

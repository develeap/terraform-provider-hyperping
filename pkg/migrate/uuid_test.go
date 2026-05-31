// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeUUID(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		wantOK bool
	}{
		{"empty", "", "", false},
		{"plain uuid hex", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000", true},
		{"prefixed safe id", "mon_abc123", "mon_abc123", true},
		{"alphanumeric", "abcDEF0123", "abcDEF0123", true},
		{"underscore", "mon_safe_123", "mon_safe_123", true},
		{"hyphen", "mon-safe-123", "mon-safe-123", true},
		// Adversarial inputs the server should never produce, but which we
		// defend against in case of a backend bug or partner-API forwarding.
		{"command sub", "$(rm -rf /)", "", false},
		{"backtick", "`whoami`", "", false},
		{"semicolon chain", "abc; rm -rf $HOME", "", false},
		{"single quote escape", "'; rm -rf $HOME; '", "", false},
		{"shell var", "$HOME", "", false},
		{"space", "abc 123", "", false},
		{"slash", "../etc/passwd", "", false},
		{"newline", "abc\nrm", "", false},
		{"hcl template", `${file("/etc/passwd")}`, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := SafeUUID(tt.input)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestQuoteShellUUID(t *testing.T) {
	// Whitelisted UUID -> Go-quoted form unchanged.
	assert.Equal(t, `"mon_abc123"`, QuoteShellUUID("mon_abc123"))
	assert.Equal(t, `"550e8400-e29b-41d4-a716-446655440000"`, QuoteShellUUID("550e8400-e29b-41d4-a716-446655440000"))

	// Adversarial values must be replaced with the sentinel so they cannot
	// expand in a bash double-quoted string.
	bad := QuoteShellUUID("$(rm -rf /)")
	assert.NotContains(t, bad, "$(")
	assert.Contains(t, bad, "INVALID_UUID")

	bad2 := QuoteShellUUID("`whoami`")
	assert.NotContains(t, bad2, "`")
	assert.Contains(t, bad2, "INVALID_UUID")

	bad3 := QuoteShellUUID("abc; rm -rf $HOME")
	assert.NotContains(t, bad3, ";")
	assert.NotContains(t, bad3, "$HOME")
	assert.Contains(t, bad3, "INVALID_UUID")
}

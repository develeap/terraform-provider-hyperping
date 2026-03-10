// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeHCL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no escaping needed", "hello world", "hello world"},
		{"backslash", `path\to\file`, `path\\to\\file`},
		{"double quotes", `say "hello"`, `say \"hello\"`},
		{"newline", "line1\nline2", `line1\nline2`},
		{"carriage return", "line1\rline2", `line1\rline2`},
		{"tab", "col1\tcol2", `col1\tcol2`},
		{"mixed", "say \"hello\"\nnext line", `say \"hello\"\nnext line`},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeHCL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeShell(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no escaping needed", "hello world", "hello world"},
		{"backslash", `path\to\file`, `path\\to\\file`},
		{"double quotes", `say "hello"`, `say \"hello\"`},
		{"dollar sign", "cost is $100", `cost is \$100`},
		{"backtick", "run `ls`", "run \\`ls\\`"},
		{"mixed", `var="$HOME"`, `var=\"\$HOME\"`},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeShell(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuoteHCL(t *testing.T) {
	assert.Equal(t, `"hello"`, QuoteHCL("hello"))
	assert.Equal(t, `"say \"hi\""`, QuoteHCL(`say "hi"`))
	assert.Equal(t, `""`, QuoteHCL(""))
}

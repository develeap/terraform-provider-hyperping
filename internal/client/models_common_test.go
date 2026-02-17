// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"strings"
	"testing"
)

// TestValidateLocalizedText tests all language paths to increase coverage from 53.8% to 100%
func TestValidateLocalizedText(t *testing.T) {
	tests := []struct {
		name        string
		text        LocalizedText
		maxLen      int
		wantErr     bool
		errContains string
	}{
		{
			name:    "empty localized text is valid",
			text:    LocalizedText{},
			maxLen:  100,
			wantErr: false,
		},
		{
			name:    "english text within limit",
			text:    LocalizedText{En: "Hello"},
			maxLen:  100,
			wantErr: false,
		},
		{
			name:        "english text exceeds limit",
			text:        LocalizedText{En: strings.Repeat("a", 101)},
			maxLen:      100,
			wantErr:     true,
			errContains: "field.en",
		},
		{
			name:    "french text within limit",
			text:    LocalizedText{Fr: "Bonjour"},
			maxLen:  100,
			wantErr: false,
		},
		{
			name:        "french text exceeds limit",
			text:        LocalizedText{Fr: strings.Repeat("b", 101)},
			maxLen:      100,
			wantErr:     true,
			errContains: "field.fr",
		},
		{
			name:    "german text within limit",
			text:    LocalizedText{De: "Guten Tag"},
			maxLen:  100,
			wantErr: false,
		},
		{
			name:        "german text exceeds limit",
			text:        LocalizedText{De: strings.Repeat("c", 101)},
			maxLen:      100,
			wantErr:     true,
			errContains: "field.de",
		},
		{
			name:    "spanish text within limit",
			text:    LocalizedText{Es: "Hola"},
			maxLen:  100,
			wantErr: false,
		},
		{
			name:        "spanish text exceeds limit",
			text:        LocalizedText{Es: strings.Repeat("d", 101)},
			maxLen:      100,
			wantErr:     true,
			errContains: "field.es",
		},
		{
			name: "multiple languages all within limit",
			text: LocalizedText{
				En: "Hello",
				Fr: "Bonjour",
				De: "Hallo",
				Es: "Hola",
			},
			maxLen:  100,
			wantErr: false,
		},
		{
			name: "multiple languages with one exceeding limit",
			text: LocalizedText{
				En: "Hello",
				Fr: strings.Repeat("x", 101),
				De: "Hallo",
				Es: "Hola",
			},
			maxLen:      100,
			wantErr:     true,
			errContains: "field.fr",
		},
		{
			name: "all languages at exact limit",
			text: LocalizedText{
				En: strings.Repeat("a", 50),
				Fr: strings.Repeat("b", 50),
				De: strings.Repeat("c", 50),
				Es: strings.Repeat("d", 50),
			},
			maxLen:  50,
			wantErr: false,
		},
		{
			name:        "validation with zero max length",
			text:        LocalizedText{En: "a"},
			maxLen:      0,
			wantErr:     true,
			errContains: "",
		},
		{
			name: "empty strings in all fields",
			text: LocalizedText{
				En: "",
				Fr: "",
				De: "",
				Es: "",
			},
			maxLen:  5,
			wantErr: false,
		},
		{
			name: "partial population with varying lengths",
			text: LocalizedText{
				En: "Short",
				De: strings.Repeat("x", 50),
			},
			maxLen:  100,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLocalizedText("field", tt.text, tt.maxLen)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
				return
			}
			if tt.wantErr && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got: %v", tt.errContains, err)
			}
		})
	}
}

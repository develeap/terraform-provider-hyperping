// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package interactive

import (
	"testing"
)

func TestAPIKeyValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "valid string",
			input:   "test-api-key",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := APIKeyValidator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIKeyValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHyperpingAPIKeyValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "valid hyperping key",
			input:   "sk_test123",
			wantErr: false,
		},
		{
			name:    "invalid prefix",
			input:   "test123",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HyperpingAPIKeyValidator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("HyperpingAPIKeyValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSourceAPIKeyValidator(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		input    interface{}
		wantErr  bool
	}{
		{
			name:     "valid betterstack key",
			platform: "betterstack",
			input:    "longapikey123",
			wantErr:  false,
		},
		{
			name:     "short betterstack key",
			platform: "betterstack",
			input:    "short",
			wantErr:  true,
		},
		{
			name:     "valid uptimerobot key",
			platform: "uptimerobot",
			input:    "u123456",
			wantErr:  false,
		},
		{
			name:     "invalid uptimerobot key prefix",
			platform: "uptimerobot",
			input:    "x123456",
			wantErr:  true,
		},
		{
			name:     "valid pingdom key",
			platform: "pingdom",
			input:    "12345678901234567890",
			wantErr:  false,
		},
		{
			name:     "short pingdom key",
			platform: "pingdom",
			input:    "short",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := SourceAPIKeyValidator(tt.platform)
			err := validator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SourceAPIKeyValidator(%s)() error = %v, wantErr %v", tt.platform, err, tt.wantErr)
			}
		})
	}
}

func TestFilePathValidator(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "valid file path",
			input:   "/path/to/file.txt",
			wantErr: false,
		},
		{
			name:    "empty path",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   123,
			wantErr: true,
		},
		{
			name:    "path with null character",
			input:   "/path/\x00/file",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FilePathValidator(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePathValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsInteractive(t *testing.T) {
	// This test just ensures the function doesn't panic
	// Actual behavior depends on terminal environment
	result := IsInteractive()
	// Result can be true or false depending on test environment
	_ = result
}

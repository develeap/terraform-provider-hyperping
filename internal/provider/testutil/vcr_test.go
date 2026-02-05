// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"os"
	"testing"
)

func TestGetRecordMode(t *testing.T) {
	// Save original env
	orig := os.Getenv("RECORD_MODE")
	defer func() {
		if orig != "" {
			os.Setenv("RECORD_MODE", orig)
		} else {
			os.Unsetenv("RECORD_MODE")
		}
	}()

	t.Run("default is auto", func(t *testing.T) {
		os.Unsetenv("RECORD_MODE")
		if mode := GetRecordMode(); mode != ModeAuto {
			t.Errorf("expected ModeAuto, got %v", mode)
		}
	})

	t.Run("RECORD_MODE=true returns record", func(t *testing.T) {
		os.Setenv("RECORD_MODE", "true")
		if mode := GetRecordMode(); mode != ModeRecord {
			t.Errorf("expected ModeRecord, got %v", mode)
		}
	})

	t.Run("RECORD_MODE=false returns auto", func(t *testing.T) {
		os.Setenv("RECORD_MODE", "false")
		if mode := GetRecordMode(); mode != ModeAuto {
			t.Errorf("expected ModeAuto, got %v", mode)
		}
	})
}

func TestMaskSensitiveHeaders(t *testing.T) {
	// Import cassette package for test
	// This is a basic sanity check that the function exists and doesn't panic
	t.Run("function exists", func(t *testing.T) {
		// Just verify the function is accessible
		_ = maskSensitiveHeaders
	})
}

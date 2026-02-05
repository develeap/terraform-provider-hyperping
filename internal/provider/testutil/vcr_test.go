// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
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
	t.Run("masks authorization header", func(t *testing.T) {
		interaction := &cassette.Interaction{
			Request: cassette.Request{
				Headers: http.Header{
					"Authorization": {"Bearer sk_secret123"},
				},
				URL: "https://api.example.com/v1/test",
			},
			Response: cassette.Response{
				Headers: http.Header{},
			},
		}

		maskSensitiveHeaders(interaction)

		auth := interaction.Request.Headers.Get("Authorization")
		if auth != "Bearer [MASKED]" {
			t.Errorf("expected masked auth header, got %s", auth)
		}
	})

	t.Run("masks set-cookie header", func(t *testing.T) {
		interaction := &cassette.Interaction{
			Request: cassette.Request{
				Headers: http.Header{},
				URL:     "https://api.example.com/v1/test",
			},
			Response: cassette.Response{
				Headers: http.Header{
					"Set-Cookie": {"session=abc123; Path=/"},
				},
			},
		}

		maskSensitiveHeaders(interaction)

		cookie := interaction.Response.Headers.Get("Set-Cookie")
		if cookie != "[MASKED]" {
			t.Errorf("expected masked Set-Cookie header, got %s", cookie)
		}
	})

	t.Run("masks api_key in URL", func(t *testing.T) {
		interaction := &cassette.Interaction{
			Request: cassette.Request{
				Headers: http.Header{},
				URL:     "https://api.example.com/v1/test?api_key=secret123",
			},
			Response: cassette.Response{
				Headers: http.Header{},
			},
		}

		maskSensitiveHeaders(interaction)

		// The masking replaces "api_key=" with "api_key=[MASKED]"
		if !strings.Contains(interaction.Request.URL, "api_key=[MASKED]") {
			t.Errorf("expected api_key=[MASKED] in URL, got: %s", interaction.Request.URL)
		}
	})

	t.Run("handles empty headers", func(t *testing.T) {
		interaction := &cassette.Interaction{
			Request: cassette.Request{
				Headers: http.Header{},
				URL:     "https://api.example.com/v1/test",
			},
			Response: cassette.Response{
				Headers: http.Header{},
			},
		}

		// Should not panic
		maskSensitiveHeaders(interaction)

		// URL should remain unchanged
		if interaction.Request.URL != "https://api.example.com/v1/test" {
			t.Errorf("URL should not be modified when no api_key present")
		}
	})
}

func TestRequireEnvForRecording(t *testing.T) {
	// Save original env
	origRecordMode := os.Getenv("RECORD_MODE")
	origTestKey := os.Getenv("TEST_VAR_FOR_VCR")
	defer func() {
		if origRecordMode != "" {
			os.Setenv("RECORD_MODE", origRecordMode)
		} else {
			os.Unsetenv("RECORD_MODE")
		}
		if origTestKey != "" {
			os.Setenv("TEST_VAR_FOR_VCR", origTestKey)
		} else {
			os.Unsetenv("TEST_VAR_FOR_VCR")
		}
	}()

	t.Run("passes when not in recording mode", func(t *testing.T) {
		os.Unsetenv("RECORD_MODE")
		os.Unsetenv("TEST_VAR_FOR_VCR")
		// Should not skip since not in record mode
		RequireEnvForRecording(t, "TEST_VAR_FOR_VCR")
	})

	t.Run("passes when env is set in recording mode", func(t *testing.T) {
		os.Setenv("RECORD_MODE", "true")
		os.Setenv("TEST_VAR_FOR_VCR", "some-value")
		// Should not skip since env is set
		RequireEnvForRecording(t, "TEST_VAR_FOR_VCR")
	})
}

func TestVCRModeConstants(t *testing.T) {
	// Verify mode constants are distinct
	if ModeReplay == ModeRecord {
		t.Error("ModeReplay should not equal ModeRecord")
	}
	if ModeRecord == ModeAuto {
		t.Error("ModeRecord should not equal ModeAuto")
	}
	if ModeReplay == ModeAuto {
		t.Error("ModeReplay should not equal ModeAuto")
	}

	// Verify expected values
	if ModeReplay != 0 {
		t.Errorf("ModeReplay should be 0, got %d", ModeReplay)
	}
	if ModeRecord != 1 {
		t.Errorf("ModeRecord should be 1, got %d", ModeRecord)
	}
	if ModeAuto != 2 {
		t.Errorf("ModeAuto should be 2, got %d", ModeAuto)
	}
}

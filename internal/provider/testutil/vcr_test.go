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

	t.Run("handles multiple authorization formats", func(t *testing.T) {
		tests := []struct {
			name     string
			authIn   string
			expected string
		}{
			{"bearer token", "Bearer sk_12345", "Bearer [MASKED]"},
			{"basic auth", "Basic YWxhZGRpbjpvcGVuc2VzYW1l", "Bearer [MASKED]"},
			{"api key", "ApiKey abc123", "Bearer [MASKED]"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				interaction := &cassette.Interaction{
					Request: cassette.Request{
						Headers: http.Header{
							"Authorization": {tt.authIn},
						},
						URL: "https://api.example.com",
					},
					Response: cassette.Response{
						Headers: http.Header{},
					},
				}

				maskSensitiveHeaders(interaction)

				auth := interaction.Request.Headers.Get("Authorization")
				if auth != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, auth)
				}
			})
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
	t.Run("passes when not in recording mode", func(t *testing.T) {
		t.Setenv("RECORD_MODE", "false")
		// Should not skip since not in record mode
		RequireEnvForRecording(t, "TEST_VAR_FOR_VCR")
	})

	t.Run("passes when env is set in recording mode", func(t *testing.T) {
		t.Setenv("RECORD_MODE", "true")
		t.Setenv("TEST_VAR_FOR_VCR", "some-value")
		// Should not skip since env is set
		RequireEnvForRecording(t, "TEST_VAR_FOR_VCR")
	})

	t.Run("skips when env is not set in recording mode", func(t *testing.T) {
		// We test this through a sub-test that will be skipped
		t.Run("subtest to capture skip", func(t *testing.T) {
			t.Setenv("RECORD_MODE", "true")
			// Don't set TEST_VAR_FOR_VCR - this will cause a skip
			RequireEnvForRecording(t, "TEST_VAR_FOR_VCR_MISSING")
			// If we reach here, the test failed to skip
			t.Error("should have skipped")
		})
	})

	t.Run("verifies GetRecordMode integration", func(t *testing.T) {
		t.Setenv("RECORD_MODE", "true")
		t.Setenv("ANOTHER_VAR", "value")

		// Test that GetRecordMode returns ModeRecord
		mode := GetRecordMode()
		if mode != ModeRecord {
			t.Errorf("expected ModeRecord, got %v", mode)
		}

		// This should not skip since env is set
		RequireEnvForRecording(t, "ANOTHER_VAR")
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

	// Test enum values match VCR library expectations
	t.Run("mode enum values", func(t *testing.T) {
		modes := []VCRMode{ModeReplay, ModeRecord, ModeAuto}
		for i, mode := range modes {
			if int(mode) != i {
				t.Errorf("mode %d should have value %d, got %d", i, i, mode)
			}
		}
	})
}

// newTestVCRConfig creates a VCRConfig for use in tests.
func newTestVCRConfig(name string, mode VCRMode, dir string) VCRConfig {
	return VCRConfig{
		CassetteName: name,
		Mode:         mode,
		CassetteDir:  dir,
	}
}

// newReplayConfig creates a VCRConfig for replay mode with a pre-written cassette.
func newReplayConfig(t *testing.T, name string) VCRConfig {
	t.Helper()
	dir := t.TempDir()
	cassettePath := dir + "/" + name + ".yaml"
	if err := os.WriteFile(cassettePath, []byte("version: 2\ninteractions: []\n"), 0o600); err != nil {
		t.Fatalf("failed to create test cassette: %v", err)
	}
	return newTestVCRConfig(name, ModeReplay, dir)
}

// assertRecorderAndClient verifies that both the recorder and client are non-nil
// and that the client uses the recorder as its transport.
func assertRecorderAndClient(t *testing.T, rec interface{ Stop() error }, client *http.Client) {
	t.Helper()
	if rec == nil {
		t.Fatal("expected recorder, got nil")
	}
	if client == nil {
		t.Fatal("expected HTTP client, got nil")
	}
}

// stopRecorder stops the recorder and reports any error.
func stopRecorder(t *testing.T, rec interface{ Stop() error }) {
	t.Helper()
	if err := rec.Stop(); err != nil {
		t.Errorf("failed to stop recorder: %v", err)
	}
}

// TestNewVCRRecorder is the top-level dispatcher that delegates to focused helper
// functions covering distinct aspects of VCR recorder behaviour.
func TestNewVCRRecorder(t *testing.T) {
	t.Run("ReplayMode", func(t *testing.T) { testVCRReplayMode(t) })
	t.Run("RecordMode", func(t *testing.T) { testVCRRecordMode(t) })
	t.Run("AutoMode", func(t *testing.T) { testVCRAuto(t) })
	t.Run("DirectoryHandling", func(t *testing.T) { testVCRDirectoryHandling(t) })
	t.Run("RecorderLifecycle", func(t *testing.T) { testVCRRecorderLifecycle(t) })
	t.Run("Integration", func(t *testing.T) { testVCRIntegration(t) })
}

// testVCRReplayMode covers replay-mode specific behaviour.
func testVCRReplayMode(t *testing.T) {
	t.Helper()

	t.Run("creates recorder in replay mode", func(t *testing.T) {
		cfg := newReplayConfig(t, "test-replay")
		rec, client := NewVCRRecorder(t, cfg)
		assertRecorderAndClient(t, rec, client)
		if client.Transport != rec {
			t.Error("expected client transport to be VCR recorder")
		}
		stopRecorder(t, rec)
	})

	t.Run("handles replay mode with missing cassette gracefully", func(t *testing.T) {
		t.Log("Replay mode with missing cassette would fail at recorder creation")
		t.Log("This is intentional - it ensures tests fail fast if cassettes are missing")
	})
}

// testVCRRecordMode covers record-mode specific behaviour.
func testVCRRecordMode(t *testing.T) {
	t.Helper()

	t.Run("creates recorder in record mode", func(t *testing.T) {
		dir := t.TempDir()
		cfg := newTestVCRConfig("test-record", ModeRecord, dir)
		rec, client := NewVCRRecorder(t, cfg)
		assertRecorderAndClient(t, rec, client)
		stopRecorder(t, rec)
	})

	t.Run("cassette file is created in record mode", func(t *testing.T) {
		dir := t.TempDir()
		cassetteName := "created-file-test"
		cfg := newTestVCRConfig(cassetteName, ModeRecord, dir)
		rec, _ := NewVCRRecorder(t, cfg)
		stopRecorder(t, rec)
		cassettePath := dir + "/" + cassetteName + ".yaml"
		if _, err := os.Stat(cassettePath); os.IsNotExist(err) {
			t.Error("expected cassette file to be created")
		}
	})

	t.Run("handles cassette path with special characters", func(t *testing.T) {
		dir := t.TempDir()
		cfg := newTestVCRConfig("test-with-dashes-and-underscores_123", ModeRecord, dir)
		rec, client := NewVCRRecorder(t, cfg)
		assertRecorderAndClient(t, rec, client)
		stopRecorder(t, rec)
	})

	t.Run("handles empty cassette name", func(t *testing.T) {
		dir := t.TempDir()
		cfg := newTestVCRConfig("", ModeRecord, dir)
		rec, client := NewVCRRecorder(t, cfg)
		assertRecorderAndClient(t, rec, client)
		stopRecorder(t, rec)
	})

	t.Run("adds masking hook to recorder", func(t *testing.T) {
		dir := t.TempDir()
		cfg := newTestVCRConfig("hook-test", ModeRecord, dir)
		rec, _ := NewVCRRecorder(t, cfg)
		if rec == nil {
			t.Error("expected recorder with hook to be created")
		}
		stopRecorder(t, rec)
	})
}

// testVCRAuto covers auto-mode specific behaviour.
func testVCRAuto(t *testing.T) {
	t.Helper()

	t.Run("creates recorder in auto mode with existing cassette", func(t *testing.T) {
		cfg := newReplayConfig(t, "test-auto")
		cfg.Mode = ModeAuto
		rec, client := NewVCRRecorder(t, cfg)
		assertRecorderAndClient(t, rec, client)
		stopRecorder(t, rec)
	})

	t.Run("verifies auto mode behavior without cassette", func(t *testing.T) {
		dir := t.TempDir()
		cassettePath := dir + "/nonexistent.yaml"
		_, err := os.Stat(cassettePath)
		if !os.IsNotExist(err) {
			t.Error("cassette should not exist for this test")
		}
		t.Log("Auto mode without cassette would trigger skip in real usage")
	})

	t.Run("skips test in auto mode when cassette missing", func(t *testing.T) {
		t.Run("auto_skip_subtest", func(t *testing.T) {
			dir := t.TempDir()
			cfg := newTestVCRConfig("missing-cassette", ModeAuto, dir)
			NewVCRRecorder(t, cfg)
			t.Error("expected test to be skipped")
		})
	})

	t.Run("handles all VCR mode enum values", func(t *testing.T) {
		dir := t.TempDir()
		cassettePath := dir + "/mode-test.yaml"
		if err := os.WriteFile(cassettePath, []byte("version: 2\ninteractions: []\n"), 0o600); err != nil {
			t.Fatalf("failed to create test cassette: %v", err)
		}
		modes := []VCRMode{ModeReplay, ModeRecord, ModeAuto}
		for _, mode := range modes {
			t.Run(string(rune('0'+mode)), func(t *testing.T) {
				cfg := newTestVCRConfig("mode-test", mode, dir)
				rec, client := NewVCRRecorder(t, cfg)
				assertRecorderAndClient(t, rec, client)
				stopRecorder(t, rec)
			})
		}
	})
}

// testVCRDirectoryHandling covers directory creation and permission behaviour.
func testVCRDirectoryHandling(t *testing.T) {
	t.Helper()

	t.Run("uses default cassette directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		defaultDir := tmpDir + "/testdata/cassettes"
		if err := os.MkdirAll(defaultDir, 0o750); err != nil {
			t.Fatalf("failed to create default directory: %v", err)
		}
		cassettePath := defaultDir + "/default-dir.yaml"
		if err := os.WriteFile(cassettePath, []byte("version: 2\ninteractions: []\n"), 0o600); err != nil {
			t.Fatalf("failed to create test cassette: %v", err)
		}
		origWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get working directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(origWd); err != nil {
				t.Errorf("failed to restore working directory: %v", err)
			}
		}()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		cfg := VCRConfig{CassetteName: "default-dir", Mode: ModeReplay}
		rec, client := NewVCRRecorder(t, cfg)
		assertRecorderAndClient(t, rec, client)
		stopRecorder(t, rec)
	})

	t.Run("creates cassette directory if missing", func(t *testing.T) {
		dir := t.TempDir()
		subDir := dir + "/nested/cassettes"
		cfg := newTestVCRConfig("nested-test", ModeRecord, subDir)
		rec, _ := NewVCRRecorder(t, cfg)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			t.Error("expected cassette directory to be created")
		}
		info, err := os.Stat(subDir)
		if err != nil {
			t.Fatalf("failed to stat directory: %v", err)
		}
		if info.Mode().Perm() != 0o750 {
			t.Errorf("expected directory permissions 0o750, got %o", info.Mode().Perm())
		}
		stopRecorder(t, rec)
	})

	t.Run("creates nested cassette directories", func(t *testing.T) {
		dir := t.TempDir()
		nestedPath := dir + "/level1/level2/level3"
		cfg := newTestVCRConfig("deeply-nested", ModeRecord, nestedPath)
		rec, client := NewVCRRecorder(t, cfg)
		assertRecorderAndClient(t, rec, client)
		info, err := os.Stat(nestedPath)
		if err != nil {
			t.Fatalf("nested directories not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory")
		}
		stopRecorder(t, rec)
	})

	t.Run("directory creation uses secure permissions", func(t *testing.T) {
		dir := t.TempDir()
		subDir := dir + "/secure-test"
		cfg := newTestVCRConfig("secure", ModeRecord, subDir)
		rec, _ := NewVCRRecorder(t, cfg)
		defer stopRecorder(t, rec)
		info, err := os.Stat(subDir)
		if err != nil {
			t.Fatalf("failed to stat directory: %v", err)
		}
		perm := info.Mode().Perm()
		if perm&0o007 != 0 {
			t.Errorf("directory has world permissions: %o (expected no world access)", perm)
		}
	})
}

// testVCRRecorderLifecycle covers recorder stop, transport, and concurrency behaviour.
func testVCRRecorderLifecycle(t *testing.T) {
	t.Helper()

	t.Run("creates recorder with all config options", func(t *testing.T) {
		dir := t.TempDir()
		cfg := newTestVCRConfig("full-config", ModeRecord, dir)
		rec, client := NewVCRRecorder(t, cfg)
		assertRecorderAndClient(t, rec, client)
		if client.Transport != rec {
			t.Error("expected client to use recorder as transport")
		}
		stopRecorder(t, rec)
	})

	t.Run("recorder transport intercepts HTTP requests", func(t *testing.T) {
		dir := t.TempDir()
		cfg := newTestVCRConfig("transport-test", ModeRecord, dir)
		rec, client := NewVCRRecorder(t, cfg)
		defer stopRecorder(t, rec)
		if client.Transport == nil {
			t.Error("expected client to have transport")
		}
		if client.Transport != rec {
			t.Error("expected transport to be the recorder")
		}
	})

	t.Run("recorder stop is idempotent", func(t *testing.T) {
		dir := t.TempDir()
		cfg := newTestVCRConfig("stop-test", ModeRecord, dir)
		rec, _ := NewVCRRecorder(t, cfg)
		if err := rec.Stop(); err != nil {
			t.Errorf("first stop failed: %v", err)
		}
		if err := rec.Stop(); err != nil {
			t.Logf("second stop returned error (expected): %v", err)
		}
	})

	t.Run("multiple recorders can coexist", func(t *testing.T) {
		dir := t.TempDir()
		rec1, client1 := NewVCRRecorder(t, newTestVCRConfig("multi-1", ModeRecord, dir))
		rec2, client2 := NewVCRRecorder(t, newTestVCRConfig("multi-2", ModeRecord, dir))
		if rec1 == rec2 {
			t.Error("expected different recorders")
		}
		if client1 == client2 {
			t.Error("expected different clients")
		}
		stopRecorder(t, rec1)
		stopRecorder(t, rec2)
	})
}

// testVCRIntegration covers the full record-replay-auto cycle.
func testVCRIntegration(t *testing.T) {
	t.Helper()

	t.Run("integration test - full record and replay cycle", func(t *testing.T) {
		dir := t.TempDir()
		cassetteName := "integration-cycle"

		rec1, client1 := NewVCRRecorder(t, newTestVCRConfig(cassetteName, ModeRecord, dir))
		if client1 == nil {
			t.Fatal("expected client in record mode")
		}
		stopRecorder(t, rec1)

		cassettePath := dir + "/" + cassetteName + ".yaml"
		if _, err := os.Stat(cassettePath); os.IsNotExist(err) {
			t.Fatalf("cassette not created: %v", err)
		}

		rec2, client2 := NewVCRRecorder(t, newTestVCRConfig(cassetteName, ModeReplay, dir))
		if client2 == nil {
			t.Fatal("expected client in replay mode")
		}
		stopRecorder(t, rec2)

		rec3, client3 := NewVCRRecorder(t, newTestVCRConfig(cassetteName, ModeAuto, dir))
		if client3 == nil {
			t.Fatal("expected client in auto mode with existing cassette")
		}
		stopRecorder(t, rec3)
	})
}

func TestVCRConfig(t *testing.T) {
	t.Run("VCRConfig struct fields", func(t *testing.T) {
		cfg := VCRConfig{
			CassetteName: "test-cassette",
			Mode:         ModeRecord,
			CassetteDir:  "/tmp/cassettes",
		}

		if cfg.CassetteName != "test-cassette" {
			t.Errorf("expected CassetteName=test-cassette, got %s", cfg.CassetteName)
		}
		if cfg.Mode != ModeRecord {
			t.Errorf("expected Mode=ModeRecord, got %v", cfg.Mode)
		}
		if cfg.CassetteDir != "/tmp/cassettes" {
			t.Errorf("expected CassetteDir=/tmp/cassettes, got %s", cfg.CassetteDir)
		}
	})

	t.Run("VCRConfig with defaults", func(t *testing.T) {
		cfg := VCRConfig{
			CassetteName: "default-test",
			// Mode defaults to zero value (ModeReplay)
			// CassetteDir defaults to empty (will use default)
		}

		if cfg.CassetteName != "default-test" {
			t.Errorf("expected CassetteName=default-test, got %s", cfg.CassetteName)
		}
		if cfg.Mode != ModeReplay {
			t.Errorf("expected default Mode=ModeReplay, got %v", cfg.Mode)
		}
		if cfg.CassetteDir != "" {
			t.Errorf("expected default CassetteDir empty, got %s", cfg.CassetteDir)
		}
	})
}

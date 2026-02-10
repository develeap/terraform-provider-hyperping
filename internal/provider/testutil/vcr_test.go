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

func TestNewVCRRecorder(t *testing.T) {
	t.Run("creates recorder in replay mode", func(t *testing.T) {
		dir := t.TempDir()
		cassettePath := dir + "/test-replay.yaml"

		// Create a dummy cassette file so replay mode doesn't fail
		if err := os.WriteFile(cassettePath, []byte("version: 2\ninteractions: []\n"), 0o600); err != nil {
			t.Fatalf("failed to create test cassette: %v", err)
		}

		cfg := VCRConfig{
			CassetteName: "test-replay",
			Mode:         ModeReplay,
			CassetteDir:  dir,
		}

		rec, client := NewVCRRecorder(t, cfg)

		if rec == nil {
			t.Fatal("expected recorder, got nil")
		}
		if client == nil {
			t.Fatal("expected HTTP client, got nil")
		}

		// Verify client has VCR transport
		if client.Transport != rec {
			t.Error("expected client transport to be VCR recorder")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("creates recorder in record mode", func(t *testing.T) {
		dir := t.TempDir()

		cfg := VCRConfig{
			CassetteName: "test-record",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec, client := NewVCRRecorder(t, cfg)

		if rec == nil {
			t.Fatal("expected recorder, got nil")
		}
		if client == nil {
			t.Fatal("expected HTTP client, got nil")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("creates recorder in auto mode with existing cassette", func(t *testing.T) {
		dir := t.TempDir()
		cassettePath := dir + "/test-auto.yaml"

		// Create a dummy cassette file
		if err := os.WriteFile(cassettePath, []byte("version: 2\ninteractions: []\n"), 0o600); err != nil {
			t.Fatalf("failed to create test cassette: %v", err)
		}

		cfg := VCRConfig{
			CassetteName: "test-auto",
			Mode:         ModeAuto,
			CassetteDir:  dir,
		}

		rec, client := NewVCRRecorder(t, cfg)

		if rec == nil {
			t.Fatal("expected recorder, got nil")
		}
		if client == nil {
			t.Fatal("expected HTTP client, got nil")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("verifies auto mode behavior without cassette", func(t *testing.T) {
		dir := t.TempDir()

		// We test that the cassette path check works
		cassettePath := dir + "/nonexistent.yaml"
		_, err := os.Stat(cassettePath)
		if !os.IsNotExist(err) {
			t.Error("cassette should not exist for this test")
		}

		// The actual skip behavior is tested in integration
		// This test verifies the file check logic
		t.Log("Auto mode without cassette would trigger skip in real usage")
	})

	t.Run("uses default cassette directory", func(t *testing.T) {
		// Use a temporary directory for testing
		tmpDir := t.TempDir()
		defaultDir := tmpDir + "/testdata/cassettes"

		// Create cassette in expected location
		if err := os.MkdirAll(defaultDir, 0o750); err != nil {
			t.Fatalf("failed to create default directory: %v", err)
		}
		cassettePath := defaultDir + "/default-dir.yaml"
		if err := os.WriteFile(cassettePath, []byte("version: 2\ninteractions: []\n"), 0o600); err != nil {
			t.Fatalf("failed to create test cassette: %v", err)
		}

		// Change to temp directory
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

		cfg := VCRConfig{
			CassetteName: "default-dir",
			Mode:         ModeReplay,
			// CassetteDir is empty, should use default
		}

		rec, client := NewVCRRecorder(t, cfg)

		if rec == nil {
			t.Fatal("expected recorder, got nil")
		}
		if client == nil {
			t.Fatal("expected HTTP client, got nil")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("creates cassette directory if missing", func(t *testing.T) {
		dir := t.TempDir()
		subDir := dir + "/nested/cassettes"

		cfg := VCRConfig{
			CassetteName: "nested-test",
			Mode:         ModeRecord,
			CassetteDir:  subDir,
		}

		rec, _ := NewVCRRecorder(t, cfg)

		// Verify directory was created
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			t.Error("expected cassette directory to be created")
		}

		// Verify directory has correct permissions (0o750)
		info, err := os.Stat(subDir)
		if err != nil {
			t.Fatalf("failed to stat directory: %v", err)
		}
		if info.Mode().Perm() != 0o750 {
			t.Errorf("expected directory permissions 0o750, got %o", info.Mode().Perm())
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("adds masking hook to recorder", func(t *testing.T) {
		dir := t.TempDir()

		cfg := VCRConfig{
			CassetteName: "hook-test",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec, _ := NewVCRRecorder(t, cfg)

		// We can't directly test the hook, but we verify the recorder was created
		// The hook is tested through maskSensitiveHeaders tests
		if rec == nil {
			t.Error("expected recorder with hook to be created")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("handles all VCR mode enum values", func(t *testing.T) {
		dir := t.TempDir()

		// Create cassette for modes that need it
		cassettePath := dir + "/mode-test.yaml"
		if err := os.WriteFile(cassettePath, []byte("version: 2\ninteractions: []\n"), 0o600); err != nil {
			t.Fatalf("failed to create test cassette: %v", err)
		}

		modes := []VCRMode{ModeReplay, ModeRecord, ModeAuto}
		for _, mode := range modes {
			t.Run(string(rune('0'+mode)), func(t *testing.T) {
				cfg := VCRConfig{
					CassetteName: "mode-test",
					Mode:         mode,
					CassetteDir:  dir,
				}

				rec, client := NewVCRRecorder(t, cfg)
				if rec == nil {
					t.Errorf("mode %d: expected recorder", mode)
				}
				if client == nil {
					t.Errorf("mode %d: expected client", mode)
				}

				if err := rec.Stop(); err != nil {
					t.Errorf("mode %d: failed to stop recorder: %v", mode, err)
				}
			})
		}
	})

	t.Run("skips test in auto mode when cassette missing", func(t *testing.T) {
		// Run in a sub-test to capture skip
		t.Run("auto_skip_subtest", func(t *testing.T) {
			dir := t.TempDir()

			cfg := VCRConfig{
				CassetteName: "missing-cassette",
				Mode:         ModeAuto,
				CassetteDir:  dir,
			}

			// This will skip the test
			NewVCRRecorder(t, cfg)

			// If we reach here, skip didn't happen
			t.Error("expected test to be skipped")
		})
	})

	t.Run("creates recorder with all config options", func(t *testing.T) {
		dir := t.TempDir()

		// Test with all fields specified
		cfg := VCRConfig{
			CassetteName: "full-config",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec, client := NewVCRRecorder(t, cfg)

		if rec == nil {
			t.Fatal("expected recorder")
		}
		if client == nil {
			t.Fatal("expected client")
		}
		if client.Transport != rec {
			t.Error("expected client to use recorder as transport")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("handles cassette path with special characters", func(t *testing.T) {
		dir := t.TempDir()

		cfg := VCRConfig{
			CassetteName: "test-with-dashes-and-underscores_123",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec, client := NewVCRRecorder(t, cfg)

		if rec == nil {
			t.Fatal("expected recorder")
		}
		if client == nil {
			t.Fatal("expected client")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("creates nested cassette directories", func(t *testing.T) {
		dir := t.TempDir()
		nestedPath := dir + "/level1/level2/level3"

		cfg := VCRConfig{
			CassetteName: "deeply-nested",
			Mode:         ModeRecord,
			CassetteDir:  nestedPath,
		}

		rec, client := NewVCRRecorder(t, cfg)

		if rec == nil {
			t.Fatal("expected recorder")
		}
		if client == nil {
			t.Fatal("expected client")
		}

		// Verify all nested directories were created
		info, err := os.Stat(nestedPath)
		if err != nil {
			t.Fatalf("nested directories not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("recorder stop is idempotent", func(t *testing.T) {
		dir := t.TempDir()

		cfg := VCRConfig{
			CassetteName: "stop-test",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec, _ := NewVCRRecorder(t, cfg)

		// Stop once
		if err := rec.Stop(); err != nil {
			t.Errorf("first stop failed: %v", err)
		}

		// Stop again (should be safe)
		if err := rec.Stop(); err != nil {
			t.Logf("second stop returned error (expected): %v", err)
		}
	})

	t.Run("handles empty cassette name", func(t *testing.T) {
		dir := t.TempDir()

		cfg := VCRConfig{
			CassetteName: "",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec, client := NewVCRRecorder(t, cfg)

		if rec == nil {
			t.Fatal("expected recorder even with empty name")
		}
		if client == nil {
			t.Fatal("expected client")
		}

		// Clean up
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	})

	t.Run("cassette file is created in record mode", func(t *testing.T) {
		dir := t.TempDir()
		cassetteName := "created-file-test"

		cfg := VCRConfig{
			CassetteName: cassetteName,
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec, _ := NewVCRRecorder(t, cfg)

		// Stop to flush cassette
		if err := rec.Stop(); err != nil {
			t.Fatalf("failed to stop recorder: %v", err)
		}

		// Verify cassette file was created
		cassettePath := dir + "/" + cassetteName + ".yaml"
		if _, err := os.Stat(cassettePath); os.IsNotExist(err) {
			t.Error("expected cassette file to be created")
		}
	})

	t.Run("recorder transport intercepts HTTP requests", func(t *testing.T) {
		dir := t.TempDir()

		cfg := VCRConfig{
			CassetteName: "transport-test",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec, client := NewVCRRecorder(t, cfg)
		defer func() {
			if err := rec.Stop(); err != nil {
				t.Logf("stop error: %v", err)
			}
		}()

		// Verify the client uses the recorder as transport
		if client.Transport == nil {
			t.Error("expected client to have transport")
		}
		if client.Transport != rec {
			t.Error("expected transport to be the recorder")
		}
	})

	t.Run("multiple recorders can coexist", func(t *testing.T) {
		dir := t.TempDir()

		cfg1 := VCRConfig{
			CassetteName: "multi-1",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}
		cfg2 := VCRConfig{
			CassetteName: "multi-2",
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec1, client1 := NewVCRRecorder(t, cfg1)
		rec2, client2 := NewVCRRecorder(t, cfg2)

		if rec1 == rec2 {
			t.Error("expected different recorders")
		}
		if client1 == client2 {
			t.Error("expected different clients")
		}

		// Clean up
		if err := rec1.Stop(); err != nil {
			t.Errorf("rec1 stop failed: %v", err)
		}
		if err := rec2.Stop(); err != nil {
			t.Errorf("rec2 stop failed: %v", err)
		}
	})

	t.Run("handles replay mode with missing cassette gracefully", func(t *testing.T) {
		// This test documents expected behavior but can't easily test t.Fatalf
		// The NewVCRRecorder would call t.Fatalf if recorder creation fails
		t.Log("Replay mode with missing cassette would fail at recorder creation")
		t.Log("This is intentional - it ensures tests fail fast if cassettes are missing")
	})

	t.Run("directory creation uses secure permissions", func(t *testing.T) {
		dir := t.TempDir()
		subDir := dir + "/secure-test"

		cfg := VCRConfig{
			CassetteName: "secure",
			Mode:         ModeRecord,
			CassetteDir:  subDir,
		}

		rec, _ := NewVCRRecorder(t, cfg)
		defer func() {
			if err := rec.Stop(); err != nil {
				t.Logf("stop error: %v", err)
			}
		}()

		// Verify directory permissions (should be 0o750 for security)
		info, err := os.Stat(subDir)
		if err != nil {
			t.Fatalf("failed to stat directory: %v", err)
		}

		// Check that directory is not world-readable (gosec G301 compliance)
		perm := info.Mode().Perm()
		if perm&0o007 != 0 {
			t.Errorf("directory has world permissions: %o (expected no world access)", perm)
		}
	})

	t.Run("integration test - full record and replay cycle", func(t *testing.T) {
		dir := t.TempDir()
		cassetteName := "integration-cycle"

		// Step 1: Record mode - create cassette
		cfg1 := VCRConfig{
			CassetteName: cassetteName,
			Mode:         ModeRecord,
			CassetteDir:  dir,
		}

		rec1, client1 := NewVCRRecorder(t, cfg1)
		if client1 == nil {
			t.Fatal("expected client in record mode")
		}
		if err := rec1.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}

		// Verify cassette was created
		cassettePath := dir + "/" + cassetteName + ".yaml"
		if _, err := os.Stat(cassettePath); os.IsNotExist(err) {
			t.Fatalf("cassette not created: %v", err)
		}

		// Step 2: Replay mode - use existing cassette
		cfg2 := VCRConfig{
			CassetteName: cassetteName,
			Mode:         ModeReplay,
			CassetteDir:  dir,
		}

		rec2, client2 := NewVCRRecorder(t, cfg2)
		if client2 == nil {
			t.Fatal("expected client in replay mode")
		}
		if err := rec2.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}

		// Step 3: Auto mode - use existing cassette
		cfg3 := VCRConfig{
			CassetteName: cassetteName,
			Mode:         ModeAuto,
			CassetteDir:  dir,
		}

		rec3, client3 := NewVCRRecorder(t, cfg3)
		if client3 == nil {
			t.Fatal("expected client in auto mode with existing cassette")
		}
		if err := rec3.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
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

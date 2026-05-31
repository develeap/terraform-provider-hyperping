// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"

	hyperping "github.com/develeap/hyperping-go"
)

func TestNewTFLogAdapter(t *testing.T) {
	adapter := NewTFLogAdapter()
	if adapter == nil {
		t.Error("NewTFLogAdapter returned nil")
	}
}

func TestTFLogAdapter_Debug(t *testing.T) {
	adapter := NewTFLogAdapter()
	ctx := context.Background()

	// Test that Debug doesn't panic
	adapter.Debug(ctx, "test message", nil)
	adapter.Debug(ctx, "test with fields", map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	})
}

func TestTFLogAdapter_ImplementsLogger(t *testing.T) {
	var _ hyperping.Logger = (*TFLogAdapter)(nil)
}

// TestTFLogAdapter_MasksAPIKey verifies that secrets emitted through the
// adapter as structured fields are redacted before reaching Terraform's log
// sink. This is the runtime safety net that justifies removing the dead
// MaskAllFieldValuesRegexes calls in provider.Configure: the masking is
// applied at every Debug call rather than relying on the framework to
// propagate a derived context that it does not actually propagate.
func TestTFLogAdapter_MasksAPIKey(t *testing.T) {
	var buf bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &buf)

	adapter := NewTFLogAdapter()
	adapter.Debug(ctx, "outgoing request", map[string]interface{}{
		"api_key":       "sk_real_secret_value_should_never_leak",
		"Authorization": "Bearer sk_another_secret",
		"method":        "GET",
	})

	logged := buf.String()
	for _, secret := range []string{
		"sk_real_secret_value_should_never_leak",
		"sk_another_secret",
	} {
		if strings.Contains(logged, secret) {
			t.Errorf("secret leaked to log output:\n%s", logged)
		}
	}
	// A redaction marker must be present so the operator can confirm masking
	// is engaged.
	if !strings.Contains(logged, "[REDACTED]") && !strings.Contains(logged, "[MASKED]") {
		t.Errorf("expected redaction marker in log output:\n%s", logged)
	}
	// Non-sensitive fields are still emitted normally.
	if !strings.Contains(logged, "GET") {
		t.Errorf("expected non-sensitive field 'GET' to remain in log output:\n%s", logged)
	}
}

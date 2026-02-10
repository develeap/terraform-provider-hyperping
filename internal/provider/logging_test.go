// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
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
	var _ client.Logger = (*TFLogAdapter)(nil)
}

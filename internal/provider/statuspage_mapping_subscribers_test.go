// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	hyperping "github.com/develeap/hyperping-go"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// TestMapSubscriberToTF tests subscriber mapping
func TestMapSubscriberToTF(t *testing.T) {
	tests := []struct {
		name  string
		input *hyperping.StatusPageSubscriber
	}{
		{
			name: "email subscriber",
			input: &hyperping.StatusPageSubscriber{
				ID:        1,
				Type:      "email",
				Value:     "user@example.com",
				Email:     testutil.Ptr("user@example.com"),
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		},
		{
			name: "sms subscriber",
			input: &hyperping.StatusPageSubscriber{
				ID:        2,
				Type:      "sms",
				Value:     "+1234567890",
				Phone:     testutil.Ptr("+1234567890"),
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		},
		{
			name: "teams subscriber",
			input: &hyperping.StatusPageSubscriber{
				ID:        3,
				Type:      "teams",
				Value:     "https://outlook.office.com/webhook/...",
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		},
		{
			name: "slack subscriber",
			input: &hyperping.StatusPageSubscriber{
				ID:           4,
				Type:         "slack",
				Value:        "#general",
				SlackChannel: testutil.Ptr("#general"),
				CreatedAt:    "2026-01-31T10:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapSubscriberToTF(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if result.ID.IsNull() {
				t.Errorf("expected non-null ID")
			}

			if result.ID.ValueInt64() != int64(tt.input.ID) {
				t.Errorf("ID: expected %d, got %d", tt.input.ID, result.ID.ValueInt64())
			}

			if result.Type.ValueString() != tt.input.Type {
				t.Errorf("Type: expected %q, got %q", tt.input.Type, result.Type.ValueString())
			}

			if result.Value.ValueString() != tt.input.Value {
				t.Errorf("Value: expected %q, got %q", tt.input.Value, result.Value.ValueString())
			}

			// Check type-specific fields
			switch tt.input.Type {
			case "email":
				if result.Email.IsNull() {
					t.Errorf("expected non-null Email for email subscriber")
				}
			case "sms":
				if result.Phone.IsNull() {
					t.Errorf("expected non-null Phone for sms subscriber")
				}
			case "slack":
				if result.SlackChannel.IsNull() {
					t.Errorf("expected non-null SlackChannel for slack subscriber")
				}
			}
		})
	}
}

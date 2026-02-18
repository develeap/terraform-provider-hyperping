// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// TestMapSettingsToTF tests settings mapping
func TestMapSettingsToTF(t *testing.T) {
	tests := []struct {
		name  string
		input client.StatusPageSettings
	}{
		{
			name: "minimal settings",
			input: client.StatusPageSettings{
				Theme:       "system",
				Font:        "Inter",
				AccentColor: "#36b27e",
				Languages:   []string{"en"},
			},
		},
		{
			name: "full settings with subscribe",
			input: client.StatusPageSettings{
				Theme:       "dark",
				Font:        "Roboto",
				AccentColor: "#0066cc",
				Languages:   []string{"en", "fr", "de"},
				Subscribe: client.StatusPageSubscribeSettings{
					Enabled: true,
					Email:   true,
					SMS:     true,
					Slack:   false,
					Teams:   false,
				},
			},
		},
		{
			name: "full settings with authentication",
			input: client.StatusPageSettings{
				Theme:       "light",
				Font:        "Inter",
				AccentColor: "#ff0000",
				Languages:   []string{"en"},
				Authentication: client.StatusPageAuthenticationSettings{
					PasswordProtection: true,
					GoogleSSO:          false,
					AllowedDomains:     []string{"example.com", "test.com"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapSettingsToTF(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if result.IsNull() {
				t.Errorf("expected non-null settings object")
				return
			}

			// Verify it's an object type
			if result.IsUnknown() {
				t.Errorf("settings object is unknown")
			}
		})
	}
}

// TestMapTFToSettings tests conversion of TF settings object to API structs
func TestMapTFToSettings(t *testing.T) {
	t.Run("null object returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		subscribe, auth := mapTFToSettings(context.Background(), types.ObjectNull(StatusPageSettingsAttrTypes()), &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if subscribe != nil {
			t.Error("expected nil subscribe for null object")
		}
		if auth != nil {
			t.Error("expected nil auth for null object")
		}
	})

	t.Run("unknown object returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		subscribe, auth := mapTFToSettings(context.Background(), types.ObjectUnknown(StatusPageSettingsAttrTypes()), &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if subscribe != nil {
			t.Error("expected nil subscribe for unknown object")
		}
		if auth != nil {
			t.Error("expected nil auth for unknown object")
		}
	})
}

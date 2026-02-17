// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestExtractSubscribeSettings verifies subscribe settings extraction from a Terraform Object.
func TestExtractSubscribeSettings(t *testing.T) {
	t.Run("all 5 fields populated", func(t *testing.T) {
		obj, diags := types.ObjectValue(SubscribeSettingsAttrTypes(), map[string]attr.Value{
			"enabled": types.BoolValue(true),
			"email":   types.BoolValue(true),
			"slack":   types.BoolValue(false),
			"teams":   types.BoolValue(true),
			"sms":     types.BoolValue(false),
		})
		if diags.HasError() {
			t.Fatalf("failed to build test object: %v", diags.Errors())
		}

		var d diag.Diagnostics
		result := extractSubscribeSettings(obj, &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result == nil {
			t.Fatal("expected non-nil subscribe settings")
		}
		checkBoolPtr(t, "enabled", result.Enabled, true)
		checkBoolPtr(t, "email", result.Email, true)
		checkBoolPtr(t, "slack", result.Slack, false)
		checkBoolPtr(t, "teams", result.Teams, true)
		checkBoolPtr(t, "sms", result.SMS, false)
	})

	t.Run("returns nil when object is null", func(t *testing.T) {
		var d diag.Diagnostics
		result := extractSubscribeSettings(types.ObjectNull(SubscribeSettingsAttrTypes()), &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result != nil {
			t.Errorf("expected nil for null object, got %+v", result)
		}
	})

	t.Run("returns nil when object is unknown", func(t *testing.T) {
		var d diag.Diagnostics
		result := extractSubscribeSettings(types.ObjectUnknown(SubscribeSettingsAttrTypes()), &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result != nil {
			t.Errorf("expected nil for unknown object, got %+v", result)
		}
	})

	t.Run("null bool fields are skipped", func(t *testing.T) {
		obj, diags := types.ObjectValue(SubscribeSettingsAttrTypes(), map[string]attr.Value{
			"enabled": types.BoolNull(),
			"email":   types.BoolValue(true),
			"slack":   types.BoolNull(),
			"teams":   types.BoolNull(),
			"sms":     types.BoolNull(),
		})
		if diags.HasError() {
			t.Fatalf("failed to build test object: %v", diags.Errors())
		}

		var d diag.Diagnostics
		result := extractSubscribeSettings(obj, &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result == nil {
			t.Fatal("expected non-nil subscribe settings")
		}
		if result.Enabled != nil {
			t.Errorf("expected enabled to be nil when null, got %v", *result.Enabled)
		}
		checkBoolPtr(t, "email", result.Email, true)
		if result.Slack != nil {
			t.Errorf("expected slack to be nil when null, got %v", *result.Slack)
		}
	})
}

// TestExtractAuthSettings verifies authentication settings extraction from a Terraform Object.
func TestExtractAuthSettings(t *testing.T) {
	t.Run("all 4 fields populated including allowed_domains list", func(t *testing.T) {
		domainsList, listDiags := types.ListValue(types.StringType, []attr.Value{
			types.StringValue("example.com"),
			types.StringValue("test.org"),
		})
		if listDiags.HasError() {
			t.Fatalf("failed to build domains list: %v", listDiags.Errors())
		}

		obj, diags := types.ObjectValue(AuthenticationSettingsAttrTypes(), map[string]attr.Value{
			"password_protection": types.BoolValue(true),
			"google_sso":          types.BoolValue(false),
			"saml_sso":            types.BoolValue(true),
			"allowed_domains":     domainsList,
		})
		if diags.HasError() {
			t.Fatalf("failed to build test object: %v", diags.Errors())
		}

		var d diag.Diagnostics
		result := extractAuthSettings(obj, &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result == nil {
			t.Fatal("expected non-nil auth settings")
		}
		checkBoolPtr(t, "password_protection", result.PasswordProtection, true)
		checkBoolPtr(t, "google_sso", result.GoogleSSO, false)
		checkBoolPtr(t, "saml_sso", result.SAMLSSO, true)
		if len(result.AllowedDomains) != 2 {
			t.Errorf("expected 2 allowed domains, got %d", len(result.AllowedDomains))
		}
		if result.AllowedDomains[0] != "example.com" {
			t.Errorf("expected first domain to be 'example.com', got %q", result.AllowedDomains[0])
		}
		if result.AllowedDomains[1] != "test.org" {
			t.Errorf("expected second domain to be 'test.org', got %q", result.AllowedDomains[1])
		}
	})

	t.Run("nil handling — returns nil when object is null", func(t *testing.T) {
		var d diag.Diagnostics
		result := extractAuthSettings(types.ObjectNull(AuthenticationSettingsAttrTypes()), &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result != nil {
			t.Errorf("expected nil for null object, got %+v", result)
		}
	})

	t.Run("nil handling — returns nil when object is unknown", func(t *testing.T) {
		var d diag.Diagnostics
		result := extractAuthSettings(types.ObjectUnknown(AuthenticationSettingsAttrTypes()), &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result != nil {
			t.Errorf("expected nil for unknown object, got %+v", result)
		}
	})

	t.Run("null bool fields are skipped", func(t *testing.T) {
		domainsList, listDiags := types.ListValue(types.StringType, []attr.Value{})
		if listDiags.HasError() {
			t.Fatalf("failed to build domains list: %v", listDiags.Errors())
		}

		obj, diags := types.ObjectValue(AuthenticationSettingsAttrTypes(), map[string]attr.Value{
			"password_protection": types.BoolNull(),
			"google_sso":          types.BoolValue(true),
			"saml_sso":            types.BoolNull(),
			"allowed_domains":     domainsList,
		})
		if diags.HasError() {
			t.Fatalf("failed to build test object: %v", diags.Errors())
		}

		var d diag.Diagnostics
		result := extractAuthSettings(obj, &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result == nil {
			t.Fatal("expected non-nil auth settings")
		}
		if result.PasswordProtection != nil {
			t.Errorf("expected password_protection to be nil when null, got %v", *result.PasswordProtection)
		}
		checkBoolPtr(t, "google_sso", result.GoogleSSO, true)
		if result.SAMLSSO != nil {
			t.Errorf("expected saml_sso to be nil when null, got %v", *result.SAMLSSO)
		}
	})

	t.Run("empty allowed_domains list produces empty slice", func(t *testing.T) {
		emptyList, _ := types.ListValue(types.StringType, []attr.Value{})

		obj, diags := types.ObjectValue(AuthenticationSettingsAttrTypes(), map[string]attr.Value{
			"password_protection": types.BoolValue(false),
			"google_sso":          types.BoolValue(false),
			"saml_sso":            types.BoolValue(false),
			"allowed_domains":     emptyList,
		})
		if diags.HasError() {
			t.Fatalf("failed to build test object: %v", diags.Errors())
		}

		var d diag.Diagnostics
		result := extractAuthSettings(obj, &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result == nil {
			t.Fatal("expected non-nil auth settings")
		}
		if len(result.AllowedDomains) != 0 {
			t.Errorf("expected empty allowed_domains, got %v", result.AllowedDomains)
		}
	})

	t.Run("null allowed_domains list produces nil slice", func(t *testing.T) {
		obj, diags := types.ObjectValue(AuthenticationSettingsAttrTypes(), map[string]attr.Value{
			"password_protection": types.BoolValue(false),
			"google_sso":          types.BoolValue(false),
			"saml_sso":            types.BoolValue(false),
			"allowed_domains":     types.ListNull(types.StringType),
		})
		if diags.HasError() {
			t.Fatalf("failed to build test object: %v", diags.Errors())
		}

		var d diag.Diagnostics
		result := extractAuthSettings(obj, &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result == nil {
			t.Fatal("expected non-nil auth settings")
		}
		if result.AllowedDomains != nil {
			t.Errorf("expected nil allowed_domains for null list, got %v", result.AllowedDomains)
		}
	})
}

// checkBoolPtr is a test helper that checks a *bool pointer equals the expected value.
func checkBoolPtr(t *testing.T, name string, got *bool, want bool) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: expected %v but got nil", name, want)
		return
	}
	if *got != want {
		t.Errorf("%s: expected %v but got %v", name, want, *got)
	}
}

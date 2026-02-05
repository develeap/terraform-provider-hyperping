// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRequiredWhenValueIs_Description(t *testing.T) {
	v := RequiredWhenValueIs(path.Root("type"), "email", "subscriber type")

	desc := v.Description(context.Background())
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(context.Background())
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestRequiredWhenProtocolPort_Description(t *testing.T) {
	v := RequiredWhenProtocolPort(path.Root("protocol"))

	desc := v.Description(context.Background())
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(context.Background())
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestAtLeastOneOf_Description(t *testing.T) {
	v := AtLeastOneOf()

	desc := v.Description(context.Background())
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(context.Background())
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestAtLeastOneOf_ValidatorInterface(t *testing.T) {
	// Verify that AtLeastOneOf returns a valid validator.String interface
	v := AtLeastOneOf()

	// Type assertion should succeed
	_, ok := v.(validator.String)
	if !ok {
		t.Error("AtLeastOneOf should return a validator.String")
	}
}

func TestNoSlackSubscriberType_Description(t *testing.T) {
	v := NoSlackSubscriberType()

	desc := v.Description(context.Background())
	if desc == "" {
		t.Error("expected non-empty description")
	}

	mdDesc := v.MarkdownDescription(context.Background())
	if mdDesc == "" {
		t.Error("expected non-empty markdown description")
	}
}

func TestNoSlackSubscriberType_ValidateString(t *testing.T) {
	tests := []struct {
		name      string
		input     types.String
		wantError bool
	}{
		{
			name:      "email type allowed",
			input:     types.StringValue("email"),
			wantError: false,
		},
		{
			name:      "sms type allowed",
			input:     types.StringValue("sms"),
			wantError: false,
		},
		{
			name:      "teams type allowed",
			input:     types.StringValue("teams"),
			wantError: false,
		},
		{
			name:      "slack type rejected",
			input:     types.StringValue("slack"),
			wantError: true,
		},
		{
			name:      "null value skipped",
			input:     types.StringNull(),
			wantError: false,
		},
		{
			name:      "unknown value skipped",
			input:     types.StringUnknown(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NoSlackSubscriberType()
			req := validator.StringRequest{
				Path:        path.Root("type"),
				ConfigValue: tt.input,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("NoSlackSubscriberType(%q): got error=%v, want error=%v",
					tt.input.ValueString(), hasError, tt.wantError)
			}
		})
	}
}

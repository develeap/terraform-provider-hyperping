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

// TestConditionalValidatorDescriptions verifies Description and MarkdownDescription
// for all conditional validators return non-empty values.
func TestConditionalValidatorDescriptions(t *testing.T) {
	t.Parallel()

	type stringVal struct {
		name      string
		validator validator.String
	}
	type int64Val struct {
		name      string
		validator validator.Int64
	}

	stringValidators := []stringVal{
		{"RequiredWhenValueIs", RequiredWhenValueIs(path.Root("type"), "email", "subscriber type")},
		{"AtLeastOneOf", AtLeastOneOf()},
		{"NoSlackSubscriberType", NoSlackSubscriberType()},
	}
	int64Validators := []int64Val{
		{"RequiredWhenProtocolPort", RequiredWhenProtocolPort(path.Root("protocol"))},
	}

	ctx := context.Background()

	for _, tt := range stringValidators {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			desc := tt.validator.Description(ctx)
			if desc == "" {
				t.Error("expected non-empty description")
			}

			mdDesc := tt.validator.MarkdownDescription(ctx)
			if mdDesc == "" {
				t.Error("expected non-empty markdown description")
			}
		})
	}

	for _, tt := range int64Validators {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			desc := tt.validator.Description(ctx)
			if desc == "" {
				t.Error("expected non-empty description")
			}

			mdDesc := tt.validator.MarkdownDescription(ctx)
			if mdDesc == "" {
				t.Error("expected non-empty markdown description")
			}
		})
	}
}

func TestAtLeastOneOf_ValidatorInterface(t *testing.T) {
	t.Parallel()

	// Verify that AtLeastOneOf returns a valid validator.
	// The return type is enforced by Go's type system, so we just
	// verify the function does not panic and returns non-nil.
	v := AtLeastOneOf()
	if v == nil {
		t.Error("AtLeastOneOf should return a non-nil validator")
	}
}

func TestNoSlackSubscriberType_ValidateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     types.String
		wantError bool
	}{
		{"email type allowed", types.StringValue("email"), false},
		{"sms type allowed", types.StringValue("sms"), false},
		{"teams type allowed", types.StringValue("teams"), false},
		{"slack type rejected", types.StringValue("slack"), true},
		{"null value skipped", types.StringNull(), false},
		{"unknown value skipped", types.StringUnknown(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := NoSlackSubscriberType()
			req := validator.StringRequest{
				Path:        path.Root("type"),
				ConfigValue: tt.input,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), req, resp)

			if resp.Diagnostics.HasError() != tt.wantError {
				t.Errorf("NoSlackSubscriberType(%v): got error=%v, want error=%v",
					tt.input, resp.Diagnostics.HasError(), tt.wantError)
			}
		})
	}
}

// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// RequiredWhenValueIs validates that an attribute is set when another attribute has a specific value
type requiredWhenValueIsValidator struct {
	pathExpr  path.Path
	value     string
	fieldName string
}

func (v requiredWhenValueIsValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("must be set when %s is %q", v.fieldName, v.value)
}

func (v requiredWhenValueIsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v requiredWhenValueIsValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the current attribute is unknown or null, we can't validate yet
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Get the value of the attribute we're checking against
	var matchValue string
	diags := req.Config.GetAttribute(ctx, v.pathExpr, &matchValue)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// If the match value equals our target value, ensure this field is set
	if matchValue == v.value {
		if req.ConfigValue.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing Required Attribute",
				fmt.Sprintf("Attribute %q must be set when %s is %q",
					req.Path.String(), v.fieldName, v.value),
			)
		}
	}
}

// RequiredWhenValueIs returns a validator that requires this attribute when another has a specific value
func RequiredWhenValueIs(pathExpr path.Path, value, fieldName string) validator.String {
	return requiredWhenValueIsValidator{
		pathExpr:  pathExpr,
		value:     value,
		fieldName: fieldName,
	}
}

// RequiredWhenProtocolPort validates that port is set when protocol is "port"
type requiredWhenProtocolPortValidator struct {
	pathExpr path.Path
}

func (v requiredWhenProtocolPortValidator) Description(ctx context.Context) string {
	return "must be set when protocol is 'port'"
}

func (v requiredWhenProtocolPortValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v requiredWhenProtocolPortValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// If port is unknown, skip validation
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Get protocol value
	var protocol string
	diags := req.Config.GetAttribute(ctx, v.pathExpr, &protocol)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// If protocol is "port", port must be set
	if protocol == "port" {
		if req.ConfigValue.IsNull() {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing Required Attribute",
				"Attribute 'port' must be set when protocol is 'port'",
			)
		}
	}
}

// RequiredWhenProtocolPort returns a validator for port field
func RequiredWhenProtocolPort(protocolPath path.Path) validator.Int64 {
	return requiredWhenProtocolPortValidator{
		pathExpr: protocolPath,
	}
}

// StatusPageHostnameOrSubdomainValidator ensures at least one is set
type statusPageHostnameOrSubdomainValidator struct{}

func (v statusPageHostnameOrSubdomainValidator) Description(ctx context.Context) string {
	return "either hostname or subdomain must be set"
}

func (v statusPageHostnameOrSubdomainValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v statusPageHostnameOrSubdomainValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// This validator is applied to both hostname and subdomain
	// We need to check if at least one is set

	// If this field is set, we're good
	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueString() != "" {
		return
	}

	// Check if this is hostname or subdomain field
	pathStr := req.Path.String()
	isHostname := pathStr == "hostname"

	var otherFieldPath path.Path

	if isHostname {
		otherFieldPath = path.Root("subdomain")
	} else {
		otherFieldPath = path.Root("hostname")
	}

	// Get the other field value
	var otherValue string
	diags := req.Config.GetAttribute(ctx, otherFieldPath, &otherValue)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// If both are empty, that's an error
	if otherValue == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Configuration",
			"Either 'hostname' or 'subdomain' must be set. Both cannot be empty.",
		)
	}
}

// AtLeastOneOf returns a validator that requires at least one of hostname or subdomain
func AtLeastOneOf() validator.String {
	return statusPageHostnameOrSubdomainValidator{}
}

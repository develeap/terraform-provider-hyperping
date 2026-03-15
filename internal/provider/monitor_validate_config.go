// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ValidateConfig implements resource.ResourceWithValidateConfig for cross-field
// validation of monitor protocol/field combinations. This runs at plan time,
// before any API call, giving users immediate feedback on invalid configurations.
func (r *MonitorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var protocol types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("protocol"), &protocol)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation when protocol is unknown or null (module composition support).
	if protocol.IsUnknown() || protocol.IsNull() {
		return
	}

	protocolValue := protocol.ValueString()

	switch protocolValue {
	case "icmp":
		validateNonHTTPProtocol(ctx, req, resp, "icmp")
		validatePortNotSet(ctx, req, resp, "icmp")
	case "port":
		validateNonHTTPProtocol(ctx, req, resp, "port")
		validatePortRequired(ctx, req, resp)
	case "http":
		validateHTTPProtocol(ctx, req, resp)
	}
}

// validateNonHTTPProtocol checks that HTTP-only fields are not set for non-HTTP protocols.
func validateNonHTTPProtocol(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, protocol string) {
	checkStringNotSet(ctx, req, resp, "http_method", protocol)
	checkStringNotSet(ctx, req, resp, "expected_status_code", protocol)
	checkBoolNotSet(ctx, req, resp, "follow_redirects", protocol)
	checkListNotSet(ctx, req, resp, "request_headers", protocol)
	checkStringNotSet(ctx, req, resp, "request_body", protocol)
	checkStringNotSet(ctx, req, resp, "required_keyword", protocol)
}

// validatePortRequired checks that port is set when protocol is "port".
func validatePortRequired(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var port types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("port"), &port)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if port.IsNull() || port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Missing Required Attribute",
			"port is required when protocol is \"port\". Specify a TCP port number (1-65535).",
		)
	}
}

// validatePortNotSet checks that port is not set for protocols that do not use it.
func validatePortNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, protocol string) {
	var port types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("port"), &port)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !port.IsNull() && !port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Invalid Attribute Combination",
			fmt.Sprintf("port is not valid when protocol is %q. Remove port or change protocol to \"port\".", protocol),
		)
	}
}

// validateHTTPProtocol checks that port is not set for HTTP monitors.
func validateHTTPProtocol(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var port types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("port"), &port)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !port.IsNull() && !port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Invalid Attribute Combination",
			"port is not valid when protocol is \"http\". The URL contains the port for HTTP monitors. Remove port or change protocol to \"port\".",
		)
	}
}

// checkStringNotSet adds a diagnostic error if a string attribute is explicitly set.
func checkStringNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, attrName, protocol string) {
	var val types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrName), &val)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !val.IsNull() && !val.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrName),
			"Invalid Attribute Combination",
			fmt.Sprintf("%s is only valid for HTTP monitors. When protocol is %q, remove %s or change protocol to \"http\".",
				attrName, protocol, attrName),
		)
	}
}

// checkBoolNotSet adds a diagnostic error if a bool attribute is explicitly set.
func checkBoolNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, attrName, protocol string) {
	var val types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrName), &val)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !val.IsNull() && !val.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrName),
			"Invalid Attribute Combination",
			fmt.Sprintf("%s is only valid for HTTP monitors. When protocol is %q, remove %s or change protocol to \"http\".",
				attrName, protocol, attrName),
		)
	}
}

// checkListNotSet adds a diagnostic error if a list attribute is explicitly set.
func checkListNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, attrName, protocol string) {
	var val types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrName), &val)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !val.IsNull() && !val.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrName),
			"Invalid Attribute Combination",
			fmt.Sprintf("%s is only valid for HTTP monitors. When protocol is %q, remove %s or change protocol to \"http\".",
				attrName, protocol, attrName),
		)
	}
}

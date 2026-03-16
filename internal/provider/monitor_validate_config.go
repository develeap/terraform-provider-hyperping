// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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
		validateURLIsHTTP(ctx, req, resp)
		validateNonHTTPProtocol(ctx, req, resp, "icmp")
		validatePortNotSet(ctx, req, resp, "icmp")
		validateDNSFieldsNotSet(ctx, req, resp, "icmp")
	case "port":
		validateURLIsHTTP(ctx, req, resp)
		validateNonHTTPProtocol(ctx, req, resp, "port")
		validatePortRequired(ctx, req, resp)
		validateDNSFieldsNotSet(ctx, req, resp, "port")
	case "http":
		validateURLIsHTTP(ctx, req, resp)
		validateHTTPProtocol(ctx, req, resp)
		validateDNSFieldsNotSet(ctx, req, resp, "http")
	case "dns":
		validateNonHTTPProtocol(ctx, req, resp, "dns")
		validatePortNotSet(ctx, req, resp, "dns")
	}
}

// validateURLIsHTTP checks that the url attribute is a valid HTTP or HTTPS URL.
// This applies to http, port, and icmp protocols. DNS monitors use bare domains instead.
func validateURLIsHTTP(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var urlVal types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("url"), &urlVal)...)
	if resp.Diagnostics.HasError() || urlVal.IsNull() || urlVal.IsUnknown() {
		return
	}

	value := urlVal.ValueString()
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Invalid URL Format",
			fmt.Sprintf("The value %q must be a valid HTTP or HTTPS URL (e.g., \"https://example.com\"). "+
				"For DNS monitors, bare domain names are accepted.", value),
		)
	}
}

// validateNonHTTPProtocol checks that HTTP-only fields are not set for non-HTTP protocols.
func validateNonHTTPProtocol(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, protocol string) {
	checkStringNotSet(ctx, req, resp, "http_method", protocol, "http")
	checkStringNotSet(ctx, req, resp, "expected_status_code", protocol, "http")
	checkBoolNotSet(ctx, req, resp, "follow_redirects", protocol)
	checkListNotSet(ctx, req, resp, "request_headers", protocol)
	checkStringNotSet(ctx, req, resp, "request_body", protocol, "http")
	checkStringNotSet(ctx, req, resp, "required_keyword", protocol, "http")
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
	checkPortNotSet(ctx, req, resp,
		fmt.Sprintf("port is not valid when protocol is %q. Remove port or change protocol to \"port\".", protocol),
	)
}

// validateHTTPProtocol checks that port is not set for HTTP monitors.
func validateHTTPProtocol(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	checkPortNotSet(ctx, req, resp,
		"port is not valid when protocol is \"http\". The URL contains the port for HTTP monitors. Remove port or change protocol to \"port\".",
	)
}

// checkPortNotSet reads the port attribute and adds an error if it is explicitly set.
// errorDetail is the full human-readable detail message to use in the diagnostic.
func checkPortNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, errorDetail string) {
	var port types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("port"), &port)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !port.IsNull() && !port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("port"),
			"Invalid Attribute Combination",
			errorDetail,
		)
	}
}

// validateDNSFieldsNotSet checks that DNS-only fields are not set for non-DNS protocols.
func validateDNSFieldsNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, protocol string) {
	checkStringNotSet(ctx, req, resp, "dns_record_type", protocol, "dns")
	checkStringNotSet(ctx, req, resp, "dns_nameserver", protocol, "dns")
	checkStringNotSet(ctx, req, resp, "dns_expected_answer", protocol, "dns")
}

// checkStringNotSet adds a diagnostic error if a string attribute is explicitly set.
// ownerProtocol is the protocol that owns this field (e.g. "http", "dns").
func checkStringNotSet(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, attrName, protocol, ownerProtocol string) {
	var val types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attrName), &val)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !val.IsNull() && !val.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrName),
			"Invalid Attribute Combination",
			fmt.Sprintf("%s is only valid for %s monitors. When protocol is %q, remove %s or change protocol to %q.",
				attrName, strings.ToUpper(ownerProtocol), protocol, attrName, ownerProtocol),
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

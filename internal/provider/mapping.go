// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// MapMonitorCommonFields maps common monitor fields from API response to Terraform types.
// This is shared between MonitorResource and MonitorsDataSource to avoid duplication.
func MapMonitorCommonFields(monitor *client.Monitor, diags *diag.Diagnostics) MonitorCommonFields {
	result := MonitorCommonFields{
		ID:                 types.StringValue(monitor.UUID),
		Name:               types.StringValue(monitor.Name),
		URL:                types.StringValue(monitor.URL),
		Protocol:           types.StringValue(monitor.Protocol),
		HTTPMethod:         types.StringValue(monitor.HTTPMethod),
		CheckFrequency:     types.Int64Value(int64(monitor.CheckFrequency)),
		ExpectedStatusCode: types.StringValue(string(monitor.ExpectedStatusCode)),
		FollowRedirects:    types.BoolValue(monitor.FollowRedirects),
		Paused:             types.BoolValue(monitor.Paused),
	}

	// Handle regions
	result.Regions = mapStringSliceToList(monitor.Regions, diags)

	// Handle request headers (convert []RequestHeader to map)
	result.RequestHeaders = mapRequestHeadersToTFList(monitor.RequestHeaders, diags)

	// Handle request body
	if monitor.RequestBody != "" {
		result.RequestBody = types.StringValue(monitor.RequestBody)
	} else {
		result.RequestBody = types.StringNull()
	}

	// Handle port
	if monitor.Port != nil {
		result.Port = types.Int64Value(int64(*monitor.Port))
	} else {
		result.Port = types.Int64Null()
	}

	// Handle alerts_wait
	if monitor.AlertsWait > 0 {
		result.AlertsWait = types.Int64Value(int64(monitor.AlertsWait))
	} else {
		result.AlertsWait = types.Int64Null()
	}

	// Handle escalation_policy
	if monitor.EscalationPolicy != nil && *monitor.EscalationPolicy != "" {
		result.EscalationPolicy = types.StringValue(*monitor.EscalationPolicy)
	} else {
		result.EscalationPolicy = types.StringNull()
	}

	// Handle required_keyword
	if monitor.RequiredKeyword != nil && *monitor.RequiredKeyword != "" {
		result.RequiredKeyword = types.StringValue(*monitor.RequiredKeyword)
	} else {
		result.RequiredKeyword = types.StringNull()
	}

	return result
}

// MonitorCommonFields contains fields shared between resource and data source models.
type MonitorCommonFields struct {
	ID                 types.String
	Name               types.String
	URL                types.String
	Protocol           types.String
	HTTPMethod         types.String
	CheckFrequency     types.Int64
	Regions            types.List
	RequestHeaders     types.List // List of objects with name/value
	RequestBody        types.String
	ExpectedStatusCode types.String
	FollowRedirects    types.Bool
	Paused             types.Bool
	Port               types.Int64
	AlertsWait         types.Int64
	EscalationPolicy   types.String
	RequiredKeyword    types.String
}

// mapStringSliceToList converts a Go string slice to a Terraform List.
func mapStringSliceToList(slice []string, diags *diag.Diagnostics) types.List {
	if len(slice) == 0 {
		return types.ListNull(types.StringType)
	}

	values := make([]attr.Value, len(slice))
	for i, v := range slice {
		values[i] = types.StringValue(v)
	}

	list, listDiags := types.ListValue(types.StringType, values)
	diags.Append(listDiags...)
	return list
}

// RequestHeaderAttrTypes returns the attribute types for request headers.
func RequestHeaderAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	}
}

// monitorReferenceAttrTypes returns the attribute types for the outage monitor nested object.
func monitorReferenceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"uuid":     types.StringType,
		"name":     types.StringType,
		"url":      types.StringType,
		"protocol": types.StringType,
	}
}

// acknowledgedByAttrTypes returns the attribute types for the outage acknowledged_by nested object.
func acknowledgedByAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"uuid":  types.StringType,
		"email": types.StringType,
		"name":  types.StringType,
	}
}

// mapRequestHeadersToTFList converts []RequestHeader to a Terraform List of objects.
func mapRequestHeadersToTFList(headers []client.RequestHeader, diags *diag.Diagnostics) types.List {
	if len(headers) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()})
	}

	values := make([]attr.Value, len(headers))
	for i, h := range headers {
		obj, objDiags := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
			"name":  types.StringValue(h.Name),
			"value": types.StringValue(h.Value),
		})
		diags.Append(objDiags...)
		values[i] = obj
	}

	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}, values)
	diags.Append(listDiags...)
	return list
}

// HealthcheckCommonFields contains the shared healthcheck fields mapped from API responses.
// Used by resource, single data source, and list data source to avoid triple duplication.
type HealthcheckCommonFields struct {
	ID               types.String
	Name             types.String
	PingURL          types.String
	Cron             types.String
	Tz               types.String
	PeriodValue      types.Int64
	PeriodType       types.String
	GracePeriodValue types.Int64
	GracePeriodType  types.String
	EscalationPolicy types.String
	IsPaused         types.Bool
	IsDown           types.Bool
	Period           types.Int64
	GracePeriod      types.Int64
	LastPing         types.String
	CreatedAt        types.String
}

// MapHealthcheckCommonFields maps a client.Healthcheck to shared typed fields.
// Returns a struct with explicit Null values for all fields if hc is nil.
func MapHealthcheckCommonFields(hc *client.Healthcheck) HealthcheckCommonFields {
	if hc == nil {
		return HealthcheckCommonFields{
			ID:               types.StringNull(),
			Name:             types.StringNull(),
			PingURL:          types.StringNull(),
			Cron:             types.StringNull(),
			Tz:               types.StringNull(),
			PeriodValue:      types.Int64Null(),
			PeriodType:       types.StringNull(),
			GracePeriodValue: types.Int64Null(),
			GracePeriodType:  types.StringNull(),
			EscalationPolicy: types.StringNull(),
			IsPaused:         types.BoolNull(),
			IsDown:           types.BoolNull(),
			Period:           types.Int64Null(),
			GracePeriod:      types.Int64Null(),
			LastPing:         types.StringNull(),
			CreatedAt:        types.StringNull(),
		}
	}
	f := HealthcheckCommonFields{
		ID:               types.StringValue(hc.UUID),
		Name:             types.StringValue(hc.Name),
		PingURL:          types.StringValue(hc.PingURL),
		IsDown:           types.BoolValue(hc.IsDown),
		IsPaused:         types.BoolValue(hc.IsPaused),
		Period:           types.Int64Value(int64(hc.Period)),
		GracePeriod:      types.Int64Value(int64(hc.GracePeriod)),
		GracePeriodValue: types.Int64Value(int64(hc.GracePeriodValue)),
		GracePeriodType:  types.StringValue(hc.GracePeriodType),
	}

	if hc.Cron != "" {
		f.Cron = types.StringValue(hc.Cron)
	} else {
		f.Cron = types.StringNull()
	}
	if hc.Tz != "" {
		f.Tz = types.StringValue(hc.Tz)
	} else {
		f.Tz = types.StringNull()
	}
	if hc.PeriodValue != nil {
		f.PeriodValue = types.Int64Value(int64(*hc.PeriodValue))
	} else {
		f.PeriodValue = types.Int64Null()
	}
	if hc.PeriodType != "" {
		f.PeriodType = types.StringValue(hc.PeriodType)
	} else {
		f.PeriodType = types.StringNull()
	}
	if hc.EscalationPolicy != nil {
		f.EscalationPolicy = types.StringValue(hc.EscalationPolicy.UUID)
	} else {
		f.EscalationPolicy = types.StringNull()
	}
	if hc.LastPing != "" {
		f.LastPing = types.StringValue(hc.LastPing)
	} else {
		f.LastPing = types.StringNull()
	}
	if hc.CreatedAt != "" {
		f.CreatedAt = types.StringValue(hc.CreatedAt)
	} else {
		f.CreatedAt = types.StringNull()
	}

	return f
}

// MapOutageNestedObjects builds the monitor and acknowledged_by nested objects from an outage.
// Returns null objects if the outage or its monitor reference is missing/empty.
func MapOutageNestedObjects(outage *client.Outage, diags *diag.Diagnostics) (types.Object, types.Object) {
	if outage == nil {
		return types.ObjectNull(monitorReferenceAttrTypes()), types.ObjectNull(acknowledgedByAttrTypes())
	}

	// Guard against zero-value MonitorReference (all empty strings from malformed API response)
	var monitorObj types.Object
	if outage.Monitor.UUID == "" && outage.Monitor.Name == "" && outage.Monitor.URL == "" {
		monitorObj = types.ObjectNull(monitorReferenceAttrTypes())
	} else {
		obj, objDiags := types.ObjectValue(monitorReferenceAttrTypes(), map[string]attr.Value{
			"uuid":     types.StringValue(outage.Monitor.UUID),
			"name":     types.StringValue(outage.Monitor.Name),
			"url":      types.StringValue(outage.Monitor.URL),
			"protocol": types.StringValue(outage.Monitor.Protocol),
		})
		diags.Append(objDiags...)
		monitorObj = obj
	}

	var ackObj types.Object
	if outage.AcknowledgedBy != nil {
		obj, ackDiags := types.ObjectValue(acknowledgedByAttrTypes(), map[string]attr.Value{
			"uuid":  types.StringValue(outage.AcknowledgedBy.UUID),
			"email": types.StringValue(outage.AcknowledgedBy.Email),
			"name":  types.StringValue(outage.AcknowledgedBy.Name),
		})
		diags.Append(ackDiags...)
		ackObj = obj
	} else {
		ackObj = types.ObjectNull(acknowledgedByAttrTypes())
	}

	return monitorObj, ackObj
}

// mapTFListToRequestHeaders converts a Terraform List to []RequestHeader.
func mapTFListToRequestHeaders(list types.List, diags *diag.Diagnostics) []client.RequestHeader {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	headers := make([]client.RequestHeader, 0, len(elements))

	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			diags.AddError("Invalid header element", "Expected object type for header element")
			continue
		}

		attrs := obj.Attributes()
		name, okName := attrs["name"].(types.String)
		if !okName {
			diags.AddError("Invalid header name", "Expected string type for header name field")
			continue
		}
		value, okValue := attrs["value"].(types.String)
		if !okValue {
			diags.AddError("Invalid header value", "Expected string type for header value field")
			continue
		}

		if !name.IsNull() && !value.IsNull() {
			headers = append(headers, client.RequestHeader{
				Name:  name.ValueString(),
				Value: value.ValueString(),
			})
		}
	}

	return headers
}

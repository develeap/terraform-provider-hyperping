// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// Filter model structs for all data sources

// MonitorFilterModel represents monitor filter criteria.
type MonitorFilterModel struct {
	NameRegex   types.String `tfsdk:"name_regex"`
	Protocol    types.String `tfsdk:"protocol"`
	Paused      types.Bool   `tfsdk:"paused"`
	Status      types.String `tfsdk:"status"`
	ProjectUUID types.String `tfsdk:"project_uuid"`
}

// IncidentFilterModel represents incident filter criteria.
type IncidentFilterModel struct {
	NameRegex types.String `tfsdk:"name_regex"`
	Status    types.String `tfsdk:"status"`   // investigating, identified, monitoring, resolved
	Severity  types.String `tfsdk:"severity"` // minor, major, critical
}

// MaintenanceFilterModel represents maintenance window filter criteria.
type MaintenanceFilterModel struct {
	NameRegex types.String `tfsdk:"name_regex"`
	Status    types.String `tfsdk:"status"` // scheduled, in_progress, completed
}

// HealthcheckFilterModel represents healthcheck filter criteria.
type HealthcheckFilterModel struct {
	NameRegex types.String `tfsdk:"name_regex"`
	Status    types.String `tfsdk:"status"`
}

// OutageFilterModel represents outage filter criteria.
type OutageFilterModel struct {
	NameRegex   types.String `tfsdk:"name_regex"`
	MonitorUUID types.String `tfsdk:"monitor_uuid"`
}

// StatusPageFilterModel represents status page filter criteria.
type StatusPageFilterModel struct {
	NameRegex types.String `tfsdk:"name_regex"`
	Hostname  types.String `tfsdk:"hostname"`
}

// NameFilterModel represents a basic name-only filter.
// Can be used for resources with just name filtering.
type NameFilterModel struct {
	NameRegex types.String `tfsdk:"name_regex"`
}

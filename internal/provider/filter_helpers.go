// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Filter schema helpers - reusable blocks for client-side filtering

// NameFilterSchema returns a basic filter block for name-based filtering.
// Suitable for most resources that have a name field.
func NameFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Filter criteria for resources",
		Attributes: map[string]schema.Attribute{
			"name_regex": schema.StringAttribute{
				Optional:    true,
				Description: "Regular expression to match resource names",
			},
		},
	}
}

// MonitorFilterSchema returns filter block for monitor data sources.
// Includes name_regex, protocol, down status, and paused status.
func MonitorFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Filter criteria for monitors",
		Attributes: map[string]schema.Attribute{
			"name_regex": schema.StringAttribute{
				Optional:    true,
				Description: "Regular expression to match monitor names",
			},
			"protocol": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by protocol (http, https, tcp, icmp, udp)",
			},
			"paused": schema.BoolAttribute{
				Optional:    true,
				Description: "Filter by paused status",
			},
		},
	}
}

// IncidentFilterSchema returns filter block for incident data sources.
// Includes name_regex, status, and severity filtering.
func IncidentFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Filter criteria for incidents",
		Attributes: map[string]schema.Attribute{
			"name_regex": schema.StringAttribute{
				Optional:    true,
				Description: "Regular expression to match incident titles",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by status (investigating, identified, monitoring, resolved)",
			},
			"severity": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by severity (minor, major, critical)",
			},
		},
	}
}

// MaintenanceFilterSchema returns filter block for maintenance window data sources.
// Includes name_regex and status filtering.
func MaintenanceFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Filter criteria for maintenance windows",
		Attributes: map[string]schema.Attribute{
			"name_regex": schema.StringAttribute{
				Optional:    true,
				Description: "Regular expression to match maintenance window titles",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by status (scheduled, in_progress, completed)",
			},
		},
	}
}

// HealthcheckFilterSchema returns filter block for healthcheck data sources.
// Includes name_regex and status filtering.
func HealthcheckFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Filter criteria for healthchecks",
		Attributes: map[string]schema.Attribute{
			"name_regex": schema.StringAttribute{
				Optional:    true,
				Description: "Regular expression to match healthcheck names",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by status",
			},
		},
	}
}

// OutageFilterSchema returns filter block for outage data sources.
// Includes name_regex and monitor UUID filtering.
func OutageFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Filter criteria for outages",
		Attributes: map[string]schema.Attribute{
			"name_regex": schema.StringAttribute{
				Optional:    true,
				Description: "Regular expression to match outage monitor names",
			},
			"monitor_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by specific monitor UUID",
			},
		},
	}
}

// StatusPageFilterSchema returns filter block for status page data sources.
// Includes name_regex and hostname filtering.
func StatusPageFilterSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Filter criteria for status pages",
		Attributes: map[string]schema.Attribute{
			"name_regex": schema.StringAttribute{
				Optional:    true,
				Description: "Regular expression to match status page names",
			},
			"hostname": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by custom hostname",
			},
		},
	}
}

// Filter matching functions

// MatchesNameRegex checks if a name matches the regex pattern.
// Returns true if pattern is null/unknown (no filter).
// Returns error if regex is invalid.
func MatchesNameRegex(name string, pattern types.String) (bool, error) {
	if isNullOrUnknown(pattern) {
		return true, nil
	}

	regex, err := regexp.Compile(pattern.ValueString())
	if err != nil {
		return false, err
	}

	return regex.MatchString(name), nil
}

// MatchesExact checks for exact string match.
// Returns true if filter is null/unknown (no filter).
func MatchesExact(value string, filter types.String) bool {
	if isNullOrUnknown(filter) {
		return true
	}
	return value == filter.ValueString()
}

// MatchesExactCaseInsensitive checks for case-insensitive exact match.
// Returns true if filter is null/unknown (no filter).
func MatchesExactCaseInsensitive(value string, filter types.String) bool {
	if isNullOrUnknown(filter) {
		return true
	}
	return strings.EqualFold(value, filter.ValueString())
}

// MatchesBool checks boolean match.
// Returns true if filter is null/unknown (no filter).
func MatchesBool(value bool, filter types.Bool) bool {
	if isNullOrUnknown(filter) {
		return true
	}
	return value == filter.ValueBool()
}

// MatchesStringSlice checks if any string in slice matches the filter value.
// Returns true if filter is null/unknown (no filter).
// Useful for filtering by tags, regions, etc.
func MatchesStringSlice(values []string, filter types.String) bool {
	if isNullOrUnknown(filter) {
		return true
	}

	filterValue := filter.ValueString()
	for _, v := range values {
		if v == filterValue {
			return true
		}
	}
	return false
}

// MatchesStringSliceCaseInsensitive checks if any string in slice matches the filter value (case-insensitive).
// Returns true if filter is null/unknown (no filter).
func MatchesStringSliceCaseInsensitive(values []string, filter types.String) bool {
	if isNullOrUnknown(filter) {
		return true
	}

	filterValue := filter.ValueString()
	for _, v := range values {
		if strings.EqualFold(v, filterValue) {
			return true
		}
	}
	return false
}

// MatchesInt64 checks for exact int64 match.
// Returns true if filter is null/unknown (no filter).
func MatchesInt64(value int64, filter types.Int64) bool {
	if isNullOrUnknown(filter) {
		return true
	}
	return value == filter.ValueInt64()
}

// MatchesInt64Range checks if value is within range [minBound, maxBound] inclusive.
// Returns true if both minBound and maxBound are null/unknown (no filter).
// If only one bound is set, checks only that bound.
func MatchesInt64Range(value int64, minBound, maxBound types.Int64) bool {
	hasMin := !isNullOrUnknown(minBound)
	hasMax := !isNullOrUnknown(maxBound)

	if !hasMin && !hasMax {
		return true
	}

	if hasMin && value < minBound.ValueInt64() {
		return false
	}

	if hasMax && value > maxBound.ValueInt64() {
		return false
	}

	return true
}

// ContainsSubstring checks if value contains the filter substring (case-insensitive).
// Returns true if filter is null/unknown (no filter).
func ContainsSubstring(value string, filter types.String) bool {
	if isNullOrUnknown(filter) {
		return true
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(filter.ValueString()))
}

// ApplyAllFilters is a helper that applies multiple filter functions and returns true only if all match.
// Returns false immediately on first non-match (short-circuit evaluation).
func ApplyAllFilters(filterFuncs ...func() bool) bool {
	for _, fn := range filterFuncs {
		if !fn() {
			return false
		}
	}
	return true
}

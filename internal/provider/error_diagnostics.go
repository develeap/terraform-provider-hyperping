// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// ValidValueReference returns a formatted reference table of valid values for
// the given resource type. Returns an empty string for unrecognized types.
// This reference is appended to error messages to help users fix validation errors.
func ValidValueReference(resourceType string) string {
	switch resourceType {
	case "Monitor":
		return buildMonitorReference()
	case "Maintenance Window":
		return buildMaintenanceReference()
	case "Incident":
		return buildIncidentReference()
	default:
		return ""
	}
}

// buildMonitorReference returns the valid value reference for Monitor resources.
func buildMonitorReference() string {
	var b strings.Builder
	b.WriteString("\nQuick Reference (valid values):\n")
	fmt.Fprintf(&b, "  protocol:             %s\n", strings.Join(client.AllowedProtocols, ", "))
	fmt.Fprintf(&b, "  http_method:          %s\n", strings.Join(client.AllowedMethods, ", "))
	fmt.Fprintf(&b, "  check_frequency:      %s\n", formatIntSlice(client.AllowedFrequencies))
	b.WriteString("  expected_status_code: Specific code (200), wildcard (2xx), or range (1xx-3xx)\n")
	fmt.Fprintf(&b, "  regions:              %s\n", strings.Join(client.AllowedRegions, ", "))
	fmt.Fprintf(&b, "  alerts_wait:          %s\n", formatAlertsWaitValues())
	return b.String()
}

// buildMaintenanceReference returns the valid value reference for Maintenance Window resources.
func buildMaintenanceReference() string {
	var b strings.Builder
	b.WriteString("\nQuick Reference (valid values):\n")
	fmt.Fprintf(&b, "  notification_option:  %s\n", strings.Join(client.AllowedNotificationOptions, ", "))
	return b.String()
}

// formatIntSlice converts a slice of ints to a comma-separated string.
func formatIntSlice(values []int) string {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = strconv.Itoa(v)
	}
	return strings.Join(strs, ", ")
}

// formatAlertsWaitValues returns the sorted list of valid alerts_wait values.
func formatAlertsWaitValues() string {
	values := make([]int, 0, len(validAlertsWaitValues))
	for v := range validAlertsWaitValues {
		values = append(values, int(v))
	}
	sort.Ints(values)
	return formatIntSlice(values)
}

// buildIncidentReference returns the valid value reference for Incident resources.
func buildIncidentReference() string {
	var b strings.Builder
	b.WriteString("\nQuick Reference (valid values):\n")
	fmt.Fprintf(&b, "  type:                 %s\n", strings.Join(client.AllowedIncidentTypes, ", "))
	return b.String()
}

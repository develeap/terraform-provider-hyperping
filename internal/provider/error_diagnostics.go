// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
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
	b.WriteString("  check_frequency:      10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400\n")
	b.WriteString("  expected_status_code: Specific code (200), wildcard (2xx), or range (1xx-3xx)\n")
	fmt.Fprintf(&b, "  regions:              %s\n", strings.Join(client.AllowedRegions, ", "))
	b.WriteString("  alerts_wait:          -1, 0, 1, 2, 3, 5, 10, 30, 60\n")
	return b.String()
}

// buildMaintenanceReference returns the valid value reference for Maintenance Window resources.
func buildMaintenanceReference() string {
	var b strings.Builder
	b.WriteString("\nQuick Reference (valid values):\n")
	fmt.Fprintf(&b, "  notification_option:  %s\n", strings.Join(client.AllowedNotificationOptions, ", "))
	return b.String()
}

// buildIncidentReference returns the valid value reference for Incident resources.
func buildIncidentReference() string {
	var b strings.Builder
	b.WriteString("\nQuick Reference (valid values):\n")
	fmt.Fprintf(&b, "  type:                 %s\n", strings.Join(client.AllowedIncidentTypes, ", "))
	return b.String()
}

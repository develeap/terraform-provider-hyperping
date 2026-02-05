// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// API base paths - exported for use in tests.
// These are the canonical API paths used by the client.
const (
	// HealthchecksBasePath is the base path for healthcheck endpoints (v2).
	HealthchecksBasePath = "/v2/healthchecks"

	// IncidentsBasePath is the base path for incident endpoints (v3).
	IncidentsBasePath = "/v3/incidents"

	// MaintenanceBasePath is the base path for maintenance window endpoints (v1).
	MaintenanceBasePath = "/v1/maintenance-windows"

	// MonitorsBasePath is the base path for monitor endpoints (v1).
	MonitorsBasePath = "/v1/monitors"

	// OutagesBasePath is the base path for outage endpoints (v2).
	OutagesBasePath = "/v2/outages"

	// ReportsBasePath is the base path for report endpoints (v2).
	ReportsBasePath = "/v2/reporting/monitor-reports"

	// StatuspagesBasePath is the base path for status page endpoints (v2).
	StatuspagesBasePath = "/v2/statuspages"
)

// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package openapi converts scraped Hyperping API parameters to OpenAPI 3.0 YAML.
package openapi

// EndpointMapping maps a Hyperping docs path to an OpenAPI operation.
type EndpointMapping struct {
	DocPath    string // e.g., "monitors/create"
	HTTPMethod string // e.g., "POST"
	OASPath    string // e.g., "/v1/monitors"
	Section    string // e.g., "monitors"
}

// Mappings enumerates every known Hyperping API documentation URL path and its
// corresponding OpenAPI operation. Add new entries here when Hyperping documents
// additional endpoints.
var Mappings = []EndpointMapping{
	// Monitors – v1
	{DocPath: "monitors/list", HTTPMethod: "GET", OASPath: "/v1/monitors", Section: "monitors"},
	{DocPath: "monitors/get", HTTPMethod: "GET", OASPath: "/v1/monitors/{uuid}", Section: "monitors"},
	{DocPath: "monitors/create", HTTPMethod: "POST", OASPath: "/v1/monitors", Section: "monitors"},
	{DocPath: "monitors/update", HTTPMethod: "PUT", OASPath: "/v1/monitors/{uuid}", Section: "monitors"},
	{DocPath: "monitors/delete", HTTPMethod: "DELETE", OASPath: "/v1/monitors/{uuid}", Section: "monitors"},

	// Status Pages – v1
	{DocPath: "statuspages/list", HTTPMethod: "GET", OASPath: "/v1/status-pages", Section: "statuspages"},
	{DocPath: "statuspages/get", HTTPMethod: "GET", OASPath: "/v1/status-pages/{uuid}", Section: "statuspages"},
	{DocPath: "statuspages/create", HTTPMethod: "POST", OASPath: "/v1/status-pages", Section: "statuspages"},
	{DocPath: "statuspages/update", HTTPMethod: "PUT", OASPath: "/v1/status-pages/{uuid}", Section: "statuspages"},
	{DocPath: "statuspages/delete", HTTPMethod: "DELETE", OASPath: "/v1/status-pages/{uuid}", Section: "statuspages"},

	// Status Page Subscribers – v1
	// Section must be "statuspages_subscribers" to match analyzer.ResourceMappings APISection.
	{DocPath: "statuspages/subscribers/list", HTTPMethod: "GET", OASPath: "/v1/status-pages/{uuid}/subscribers", Section: "statuspages_subscribers"},
	{DocPath: "statuspages/subscribers/create", HTTPMethod: "POST", OASPath: "/v1/status-pages/{uuid}/subscribers", Section: "statuspages_subscribers"},
	{DocPath: "statuspages/subscribers/delete", HTTPMethod: "DELETE", OASPath: "/v1/status-pages/{uuid}/subscribers/{id}", Section: "statuspages_subscribers"},

	// Incidents – v3
	{DocPath: "incidents/list", HTTPMethod: "GET", OASPath: "/v3/incidents", Section: "incidents"},
	{DocPath: "incidents/get", HTTPMethod: "GET", OASPath: "/v3/incidents/{id}", Section: "incidents"},
	{DocPath: "incidents/create", HTTPMethod: "POST", OASPath: "/v3/incidents", Section: "incidents"},
	{DocPath: "incidents/update", HTTPMethod: "PUT", OASPath: "/v3/incidents/{id}", Section: "incidents"},
	{DocPath: "incidents/delete", HTTPMethod: "DELETE", OASPath: "/v3/incidents/{id}", Section: "incidents"},
	{DocPath: "incidents/add-update", HTTPMethod: "POST", OASPath: "/v3/incidents/{id}/updates", Section: "incidents"},

	// Maintenance Windows – v1
	{DocPath: "maintenance/list", HTTPMethod: "GET", OASPath: "/v1/maintenance-windows", Section: "maintenance"},
	{DocPath: "maintenance/get", HTTPMethod: "GET", OASPath: "/v1/maintenance-windows/{id}", Section: "maintenance"},
	{DocPath: "maintenance/create", HTTPMethod: "POST", OASPath: "/v1/maintenance-windows", Section: "maintenance"},
	{DocPath: "maintenance/update", HTTPMethod: "PUT", OASPath: "/v1/maintenance-windows/{id}", Section: "maintenance"},
	{DocPath: "maintenance/delete", HTTPMethod: "DELETE", OASPath: "/v1/maintenance-windows/{id}", Section: "maintenance"},

	// Health Checks (cron) – v1
	{DocPath: "healthchecks/list", HTTPMethod: "GET", OASPath: "/v1/health-checks", Section: "healthchecks"},
	{DocPath: "healthchecks/get", HTTPMethod: "GET", OASPath: "/v1/health-checks/{uuid}", Section: "healthchecks"},
	{DocPath: "healthchecks/create", HTTPMethod: "POST", OASPath: "/v1/health-checks", Section: "healthchecks"},
	{DocPath: "healthchecks/update", HTTPMethod: "PUT", OASPath: "/v1/health-checks/{uuid}", Section: "healthchecks"},
	{DocPath: "healthchecks/delete", HTTPMethod: "DELETE", OASPath: "/v1/health-checks/{uuid}", Section: "healthchecks"},
	{DocPath: "healthchecks/pause", HTTPMethod: "POST", OASPath: "/v1/health-checks/{uuid}/pause", Section: "healthchecks"},
	{DocPath: "healthchecks/resume", HTTPMethod: "POST", OASPath: "/v1/health-checks/{uuid}/resume", Section: "healthchecks"},

	// Outages – v1
	{DocPath: "outages/list", HTTPMethod: "GET", OASPath: "/v1/outages", Section: "outages"},
	{DocPath: "outages/get", HTTPMethod: "GET", OASPath: "/v1/outages/{id}", Section: "outages"},
	{DocPath: "outages/create", HTTPMethod: "POST", OASPath: "/v1/outages", Section: "outages"},
	{DocPath: "outages/acknowledge", HTTPMethod: "POST", OASPath: "/v1/outages/{id}/acknowledge", Section: "outages"},
	{DocPath: "outages/unacknowledge", HTTPMethod: "POST", OASPath: "/v1/outages/{id}/unacknowledge", Section: "outages"},
	{DocPath: "outages/resolve", HTTPMethod: "POST", OASPath: "/v1/outages/{id}/resolve", Section: "outages"},
	{DocPath: "outages/escalate", HTTPMethod: "POST", OASPath: "/v1/outages/{id}/escalate", Section: "outages"},
	{DocPath: "outages/delete", HTTPMethod: "DELETE", OASPath: "/v1/outages/{id}", Section: "outages"},

	// Reports – v2
	{DocPath: "reports/list", HTTPMethod: "GET", OASPath: "/v2/reporting/monitor-reports", Section: "reports"},
	{DocPath: "reports/get", HTTPMethod: "GET", OASPath: "/v2/reporting/monitor-reports/{id}", Section: "reports"},
}

// LookupByDocPath returns the mapping for the given doc path, or nil if unknown.
func LookupByDocPath(docPath string) *EndpointMapping {
	for i := range Mappings {
		if Mappings[i].DocPath == docPath {
			return &Mappings[i]
		}
	}
	return nil
}

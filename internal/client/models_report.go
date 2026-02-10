// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

// =============================================================================
// Monitor Report Models
// =============================================================================

// MonitorReport represents a report for a monitor.
// API: GET /v2/reporting/monitor-reports/{uuid}
// API: GET /v2/reporting/monitor-reports?from=&to=
type MonitorReport struct {
	UUID          string       `json:"uuid"`
	Name          string       `json:"name"`
	Protocol      string       `json:"protocol"`
	Period        ReportPeriod `json:"period"`
	SLA           float64      `json:"sla"`
	MTTR          int          `json:"mttr,omitempty"`          // Mean Time To Recovery in seconds
	MTTRFormatted string       `json:"mttrFormatted,omitempty"` // Human-readable MTTR
	Outages       OutageStats  `json:"outages"`
}

// ReportPeriod represents the time period for a report.
type ReportPeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// OutageStats contains statistics about outages.
type OutageStats struct {
	Count                  int            `json:"count"`
	TotalDowntime          int            `json:"totalDowntime"`
	TotalDowntimeFormatted string         `json:"totalDowntimeFormatted"`
	LongestOutage          int            `json:"longestOutage"`
	LongestOutageFormatted string         `json:"longestOutageFormatted"`
	Details                []OutageDetail `json:"details"`
}

// OutageDetail contains details about a specific outage.
type OutageDetail struct {
	StartDate         string `json:"startDate"`
	EndDate           string `json:"endDate"`
	Duration          int    `json:"duration,omitempty"`          // Duration in seconds
	DurationFormatted string `json:"durationFormatted,omitempty"` // Human-readable duration
}

// ListMonitorReportsResponse wraps the list response from the API.
// API: GET /v2/reporting/monitor-reports
// Response format: {"period": {...}, "monitors": [...]}
type ListMonitorReportsResponse struct {
	Period   ReportPeriod    `json:"period"`
	Monitors []MonitorReport `json:"monitors"`
}

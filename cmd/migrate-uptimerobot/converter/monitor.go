// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"fmt"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
	"github.com/develeap/terraform-provider-hyperping/pkg/migrate"
)

// HyperpingMonitor represents a Hyperping monitor configuration.
type HyperpingMonitor struct {
	ResourceName       string
	Name               string
	URL                string
	Protocol           string
	HTTPMethod         string
	CheckFrequency     int
	ExpectedStatusCode string
	RequiredKeyword    string
	Port               int
	FollowRedirects    bool
	Regions            []string
	OriginalID         int
	Warnings           []string
}

// HyperpingHealthcheck represents a Hyperping healthcheck configuration.
type HyperpingHealthcheck struct {
	ResourceName     string
	Name             string
	PeriodValue      int
	PeriodType       string
	GracePeriodValue int
	GracePeriodType  string
	OriginalID       int
	Warnings         []string
}

// ConversionResult holds the results of converting UptimeRobot monitors.
type ConversionResult struct {
	Monitors     []HyperpingMonitor
	Healthchecks []HyperpingHealthcheck
	Skipped      []SkippedMonitor
	ContactsMap  map[string][]string // Alert contact ID to list of emails/webhooks
}

// SkippedMonitor represents a monitor that couldn't be converted.
type SkippedMonitor struct {
	ID     int
	Name   string
	Type   int
	Reason string
}

// Converter converts UptimeRobot monitors to Hyperping resources.
type Converter struct{}

// NewConverter creates a new converter.
func NewConverter() *Converter {
	return &Converter{}
}

// Convert converts UptimeRobot monitors to Hyperping resources.
func (c *Converter) Convert(monitors []uptimerobot.Monitor, alertContacts []uptimerobot.AlertContact) *ConversionResult {
	result := &ConversionResult{
		Monitors:     []HyperpingMonitor{},
		Healthchecks: []HyperpingHealthcheck{},
		Skipped:      []SkippedMonitor{},
		ContactsMap:  make(map[string][]string),
	}

	// Build alert contacts map
	for _, ac := range alertContacts {
		result.ContactsMap[ac.ID] = append(result.ContactsMap[ac.ID], ac.Value)
	}

	// Convert each monitor
	seen := make(map[string]int)
	for _, m := range monitors {
		switch m.Type {
		case 1: // HTTP/HTTPS
			monitor := c.convertHTTPMonitor(m)
			monitor.ResourceName = deduplicateResourceName(monitor.ResourceName, seen)
			result.Monitors = append(result.Monitors, monitor)

		case 2: // Keyword
			monitor := c.convertKeywordMonitor(m)
			monitor.ResourceName = deduplicateResourceName(monitor.ResourceName, seen)
			result.Monitors = append(result.Monitors, monitor)

		case 3: // Ping (ICMP)
			monitor := c.convertPingMonitor(m)
			monitor.ResourceName = deduplicateResourceName(monitor.ResourceName, seen)
			result.Monitors = append(result.Monitors, monitor)

		case 4: // Port
			monitor := c.convertPortMonitor(m)
			monitor.ResourceName = deduplicateResourceName(monitor.ResourceName, seen)
			result.Monitors = append(result.Monitors, monitor)

		case 5: // Heartbeat
			healthcheck := c.convertHeartbeatMonitor(m)
			healthcheck.ResourceName = deduplicateResourceName(healthcheck.ResourceName, seen)
			result.Healthchecks = append(result.Healthchecks, healthcheck)

		default:
			result.Skipped = append(result.Skipped, SkippedMonitor{
				ID:     m.ID,
				Name:   m.FriendlyName,
				Type:   m.Type,
				Reason: fmt.Sprintf("unsupported monitor type: %d", m.Type),
			})
		}
	}

	return result
}

// convertHTTPMonitor converts an HTTP/HTTPS monitor.
func (c *Converter) convertHTTPMonitor(m uptimerobot.Monitor) HyperpingMonitor {
	monitor := HyperpingMonitor{
		ResourceName:       terraformName(m.FriendlyName),
		Name:               m.FriendlyName,
		URL:                m.URL,
		Protocol:           "http",
		HTTPMethod:         convertHTTPMethod(m.HTTPMethod),
		CheckFrequency:     mapFrequency(m.Interval),
		ExpectedStatusCode: "2xx",
		FollowRedirects:    true,
		Regions:            []string{"london", "virginia", "singapore"},
		OriginalID:         m.ID,
		Warnings:           []string{},
	}

	// Warn if frequency was adjusted
	if monitor.CheckFrequency != m.Interval {
		monitor.Warnings = append(monitor.Warnings,
			fmt.Sprintf("Check frequency adjusted from %ds to %ds (nearest allowed value)",
				m.Interval, monitor.CheckFrequency))
	}

	return monitor
}

// convertKeywordMonitor converts a keyword monitor.
func (c *Converter) convertKeywordMonitor(m uptimerobot.Monitor) HyperpingMonitor {
	monitor := HyperpingMonitor{
		ResourceName:       terraformName(m.FriendlyName),
		Name:               m.FriendlyName,
		URL:                m.URL,
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     mapFrequency(m.Interval),
		ExpectedStatusCode: "200",
		FollowRedirects:    true,
		Regions:            []string{"london", "virginia"},
		OriginalID:         m.ID,
		Warnings:           []string{},
	}

	// Handle keyword
	if m.KeywordValue != nil && *m.KeywordValue != "" {
		if m.KeywordType != nil && int(*m.KeywordType) == 2 {
			// Keyword "not exists" - not supported by Hyperping
			monitor.Warnings = append(monitor.Warnings,
				"Keyword check 'must not exist' is not supported by Hyperping. Consider using status code checks instead.")
		} else {
			// Keyword "exists" - supported
			monitor.RequiredKeyword = *m.KeywordValue
		}
	}

	// Warn if frequency was adjusted
	if monitor.CheckFrequency != m.Interval {
		monitor.Warnings = append(monitor.Warnings,
			fmt.Sprintf("Check frequency adjusted from %ds to %ds (nearest allowed value)",
				m.Interval, monitor.CheckFrequency))
	}

	return monitor
}

// convertPingMonitor converts a ping (ICMP) monitor.
func (c *Converter) convertPingMonitor(m uptimerobot.Monitor) HyperpingMonitor {
	monitor := HyperpingMonitor{
		ResourceName:   terraformName(m.FriendlyName),
		Name:           m.FriendlyName,
		URL:            ensureURLScheme(m.URL),
		Protocol:       "icmp",
		CheckFrequency: mapFrequency(m.Interval),
		Regions:        []string{"london", "virginia"},
		OriginalID:     m.ID,
		Warnings:       []string{},
	}

	// Warn if frequency was adjusted
	if monitor.CheckFrequency != m.Interval {
		monitor.Warnings = append(monitor.Warnings,
			fmt.Sprintf("Check frequency adjusted from %ds to %ds (nearest allowed value)",
				m.Interval, monitor.CheckFrequency))
	}

	return monitor
}

// convertPortMonitor converts a port monitor.
func (c *Converter) convertPortMonitor(m uptimerobot.Monitor) HyperpingMonitor {
	monitor := HyperpingMonitor{
		ResourceName:   terraformName(m.FriendlyName),
		Name:           m.FriendlyName,
		URL:            ensureURLScheme(m.URL),
		Protocol:       "port",
		CheckFrequency: mapFrequency(m.Interval),
		Regions:        []string{"virginia"},
		OriginalID:     m.ID,
		Warnings:       []string{},
	}

	// Set port from monitor configuration
	if m.Port != nil {
		monitor.Port = int(*m.Port)
	} else if m.SubType != nil {
		// Map sub-type to default port
		monitor.Port = mapSubTypeToPort(int(*m.SubType))
	} else {
		monitor.Port = 80 // Default
	}

	// Warn if frequency was adjusted
	if monitor.CheckFrequency != m.Interval {
		monitor.Warnings = append(monitor.Warnings,
			fmt.Sprintf("Check frequency adjusted from %ds to %ds (nearest allowed value)",
				m.Interval, monitor.CheckFrequency))
	}

	return monitor
}

// convertHeartbeatMonitor converts a heartbeat monitor to a healthcheck.
func (c *Converter) convertHeartbeatMonitor(m uptimerobot.Monitor) HyperpingHealthcheck {
	healthcheck := HyperpingHealthcheck{
		ResourceName:     terraformName(m.FriendlyName),
		Name:             m.FriendlyName,
		GracePeriodValue: 1,
		GracePeriodType:  "hours",
		OriginalID:       m.ID,
		Warnings:         []string{},
	}

	// Convert interval to period
	seconds := m.Interval
	if seconds >= 86400 {
		// Days
		healthcheck.PeriodValue = seconds / 86400
		healthcheck.PeriodType = "days"
	} else if seconds >= 3600 {
		// Hours
		healthcheck.PeriodValue = seconds / 3600
		healthcheck.PeriodType = "hours"
	} else if seconds >= 60 {
		// Minutes
		healthcheck.PeriodValue = seconds / 60
		healthcheck.PeriodType = "minutes"
	} else {
		// Seconds (fallback)
		healthcheck.PeriodValue = seconds
		healthcheck.PeriodType = "seconds"
	}

	healthcheck.Warnings = append(healthcheck.Warnings,
		"Heartbeat monitor converted to healthcheck. Update your script to ping the new URL (see manual-steps.md)")

	return healthcheck
}

// convertHTTPMethod converts UptimeRobot HTTP method to string.
func convertHTTPMethod(method *uptimerobot.FlexibleInt) string {
	if method == nil {
		return "GET"
	}

	switch int(*method) {
	case 1:
		return "GET"
	case 2:
		return "POST"
	case 3:
		return "PUT"
	case 4:
		return "PATCH"
	case 5:
		return "DELETE"
	case 6:
		return "HEAD"
	default:
		return "GET"
	}
}

// mapFrequency maps UptimeRobot interval to nearest Hyperping check_frequency.
func mapFrequency(interval int) int {
	return migrate.MapFrequency(interval)
}

// mapSubTypeToPort maps UptimeRobot port sub-type to default port number.
func mapSubTypeToPort(subType int) int {
	portMap := map[int]int{
		1: 80,  // Custom
		2: 80,  // HTTP
		3: 443, // HTTPS
		4: 21,  // FTP
		5: 25,  // SMTP
		6: 110, // POP3
		7: 143, // IMAP
	}

	if port, ok := portMap[subType]; ok {
		return port
	}
	return 80 // Default
}

// terraformName converts a string to a valid Terraform resource name.
// Uses "r_" prefix for digit-leading names (UptimeRobot convention).
func terraformName(name string) string {
	return migrate.SanitizeResourceNameWith(name, migrate.SanitizeOpts{
		DigitPrefix:   "r",
		EmptyFallback: "monitor",
	})
}

// ensureURLScheme prepends "https://" if the URL has no scheme.
func ensureURLScheme(rawURL string) string {
	return migrate.EnsureURLScheme(rawURL)
}

// deduplicateResourceName appends a numeric suffix when a name has already been used.
func deduplicateResourceName(name string, seen map[string]int) string {
	return migrate.DeduplicateResourceName(name, seen)
}

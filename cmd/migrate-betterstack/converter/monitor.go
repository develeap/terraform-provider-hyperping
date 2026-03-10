// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"fmt"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
	"github.com/develeap/terraform-provider-hyperping/pkg/migrate"
)

// Converter handles conversion from Better Stack to Hyperping format.
type Converter struct {
	// frequencyMap provides BetterStack-specific overrides where the desired
	// mapping differs from migrate.MapFrequency's nearest-match behavior.
	frequencyMap map[int]int
	protocolMap  map[string]string
}

// New creates a new converter with default mappings.
func New() *Converter {
	return &Converter{
		frequencyMap: map[int]int{
			45:  60,  // BetterStack rounds up to 60s (MapFrequency would pick 30s)
			240: 300, // BetterStack rounds up to 5min (MapFrequency would pick 3min)
		},
		protocolMap: map[string]string{
			"status":    "http",
			"tcp":       "port",
			"ping":      "icmp",
			"keyword":   "http", // Keyword monitors become HTTP with notes
			"heartbeat": "healthcheck",
		},
	}
}

// ConvertedMonitor represents a monitor converted to Hyperping format.
type ConvertedMonitor struct {
	ResourceName       string
	Name               string
	URL                string
	Protocol           string
	HTTPMethod         string
	CheckFrequency     int
	Regions            []string
	RequestHeaders     []RequestHeader
	RequestBody        string
	ExpectedStatusCode string
	FollowRedirects    bool
	Paused             bool
	Port               int
	Issues             []string
}

// RequestHeader represents an HTTP request header.
type RequestHeader struct {
	Name  string
	Value string
}

// ConvertedHealthcheck represents a healthcheck converted to Hyperping format.
type ConvertedHealthcheck struct {
	ResourceName string
	Name         string
	Period       int
	Grace        int
	Paused       bool
	Issues       []string
}

// ConversionIssue represents an issue encountered during conversion.
type ConversionIssue struct {
	ResourceName string
	ResourceType string
	Severity     string // "warning" or "error"
	Message      string
}

// ConvertMonitors converts Better Stack monitors to Hyperping format.
func (c *Converter) ConvertMonitors(monitors []betterstack.Monitor) ([]ConvertedMonitor, []ConversionIssue) {
	var converted []ConvertedMonitor
	var issues []ConversionIssue
	seen := make(map[string]int)

	for _, m := range monitors {
		cm, monitorIssues := c.convertMonitor(m)
		cm.ResourceName = deduplicateResourceName(cm.ResourceName, seen)
		converted = append(converted, cm)
		issues = append(issues, monitorIssues...)
	}

	return converted, issues
}

func (c *Converter) convertMonitor(m betterstack.Monitor) (ConvertedMonitor, []ConversionIssue) {
	attrs := m.Attributes
	resourceName := sanitizeResourceName(attrs.PronouncableName)
	var issues []ConversionIssue

	// Map protocol
	protocol := c.mapProtocol(attrs.MonitorType)
	if protocol == "" {
		protocol = "http"
		issues = append(issues, ConversionIssue{
			ResourceName: resourceName,
			ResourceType: "monitor",
			Severity:     "warning",
			Message:      fmt.Sprintf("Unknown monitor type '%s', defaulting to 'http'", attrs.MonitorType),
		})
	}

	// Special handling for keyword monitors
	if attrs.MonitorType == "keyword" {
		issues = append(issues, ConversionIssue{
			ResourceName: resourceName,
			ResourceType: "monitor",
			Severity:     "warning",
			Message:      "Keyword monitoring not fully supported in Hyperping. Using HTTP protocol with expected status code validation. Review required_keyword field manually.",
		})
	}

	// Map check frequency
	frequency := c.mapFrequency(attrs.CheckFrequency)
	if frequency != attrs.CheckFrequency {
		issues = append(issues, ConversionIssue{
			ResourceName: resourceName,
			ResourceType: "monitor",
			Severity:     "warning",
			Message:      fmt.Sprintf("Check frequency %ds rounded to nearest supported value %ds", attrs.CheckFrequency, frequency),
		})
	}

	// Map regions
	regions := migrate.MapRegions(attrs.Regions)
	if len(regions) == 0 {
		regions = []string{"london", "virginia", "singapore"} // Default regions
		issues = append(issues, ConversionIssue{
			ResourceName: resourceName,
			ResourceType: "monitor",
			Severity:     "warning",
			Message:      "No regions specified, using default regions: london, virginia, singapore",
		})
	}

	// Convert headers
	var headers []RequestHeader
	for _, h := range attrs.RequestHeaders {
		headers = append(headers, RequestHeader{
			Name:  h.Name,
			Value: h.Value,
		})
	}

	// Map expected status code
	expectedStatus := "200"
	if len(attrs.ExpectedStatusCodes) > 0 {
		expectedStatus = fmt.Sprintf("%d", attrs.ExpectedStatusCodes[0])
		if len(attrs.ExpectedStatusCodes) > 1 {
			issues = append(issues, ConversionIssue{
				ResourceName: resourceName,
				ResourceType: "monitor",
				Severity:     "warning",
				Message:      fmt.Sprintf("Multiple expected status codes not supported. Using first: %d. Original: %v", attrs.ExpectedStatusCodes[0], attrs.ExpectedStatusCodes),
			})
		}
	}

	// HTTP method
	method := attrs.RequestMethod
	if method == "" {
		method = "GET"
	}

	return ConvertedMonitor{
		ResourceName:       resourceName,
		Name:               attrs.PronouncableName,
		URL:                attrs.URL,
		Protocol:           protocol,
		HTTPMethod:         method,
		CheckFrequency:     frequency,
		Regions:            regions,
		RequestHeaders:     headers,
		RequestBody:        attrs.RequestBody,
		ExpectedStatusCode: expectedStatus,
		FollowRedirects:    attrs.FollowRedirects,
		Paused:             attrs.Paused,
		Port:               attrs.Port,
		Issues:             extractIssueMessages(issues),
	}, issues
}

// ConvertHeartbeats converts Better Stack heartbeats to Hyperping healthchecks.
func (c *Converter) ConvertHeartbeats(heartbeats []betterstack.Heartbeat) ([]ConvertedHealthcheck, []ConversionIssue) {
	var converted []ConvertedHealthcheck
	var issues []ConversionIssue
	seen := make(map[string]int)

	for _, h := range heartbeats {
		ch, heartbeatIssues := c.convertHeartbeat(h)
		ch.ResourceName = deduplicateResourceName(ch.ResourceName, seen)
		converted = append(converted, ch)
		issues = append(issues, heartbeatIssues...)
	}

	return converted, issues
}

func (c *Converter) convertHeartbeat(h betterstack.Heartbeat) (ConvertedHealthcheck, []ConversionIssue) {
	attrs := h.Attributes
	resourceName := sanitizeResourceName(attrs.Name)
	var issues []ConversionIssue

	// Map period to supported value
	period := c.mapFrequency(attrs.Period)
	if period != attrs.Period {
		issues = append(issues, ConversionIssue{
			ResourceName: resourceName,
			ResourceType: "healthcheck",
			Severity:     "warning",
			Message:      fmt.Sprintf("Period %ds rounded to nearest supported value %ds", attrs.Period, period),
		})
	}

	// Validate grace period
	if attrs.Grace < 60 {
		issues = append(issues, ConversionIssue{
			ResourceName: resourceName,
			ResourceType: "healthcheck",
			Severity:     "warning",
			Message:      fmt.Sprintf("Grace period %ds is less than minimum 60s. Review and adjust manually.", attrs.Grace),
		})
	}

	return ConvertedHealthcheck{
		ResourceName: resourceName,
		Name:         attrs.Name,
		Period:       period,
		Grace:        attrs.Grace,
		Paused:       attrs.Paused,
		Issues:       extractIssueMessages(issues),
	}, issues
}

func (c *Converter) mapProtocol(bsType string) string {
	if protocol, ok := c.protocolMap[bsType]; ok {
		return protocol
	}
	return ""
}

func (c *Converter) mapRegions(bsRegions []string) []string {
	return migrate.MapRegions(bsRegions)
}

func (c *Converter) mapFrequency(frequency int) int {
	if mapped, ok := c.frequencyMap[frequency]; ok {
		return mapped
	}
	return migrate.MapFrequency(frequency)
}

func sanitizeResourceName(name string) string {
	return migrate.SanitizeResourceName(name)
}

func deduplicateResourceName(name string, seen map[string]int) string {
	return migrate.DeduplicateResourceName(name, seen)
}

func extractIssueMessages(issues []ConversionIssue) []string {
	var messages []string
	for _, issue := range issues {
		messages = append(messages, issue.Message)
	}
	return messages
}

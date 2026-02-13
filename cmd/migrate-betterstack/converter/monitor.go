// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
)

// Converter handles conversion from Better Stack to Hyperping format.
type Converter struct {
	regionMap    map[string]string
	frequencyMap map[int]int
	protocolMap  map[string]string
}

// New creates a new converter with default mappings.
func New() *Converter {
	return &Converter{
		regionMap: map[string]string{
			"us":             "virginia",
			"us-east":        "virginia",
			"us-east-1":      "virginia",
			"us-west":        "oregon",
			"us-west-1":      "oregon",
			"eu":             "london",
			"eu-west":        "london",
			"eu-west-1":      "london",
			"eu-central":     "frankfurt",
			"eu-central-1":   "frankfurt",
			"asia":           "singapore",
			"ap-southeast":   "singapore",
			"ap-southeast-1": "singapore",
			"ap-northeast":   "tokyo",
			"ap-northeast-1": "tokyo",
			"au":             "sydney",
			"au-southeast":   "sydney",
			"sa":             "saopaulo",
			"sa-east-1":      "saopaulo",
		},
		frequencyMap: map[int]int{
			10:    10,
			20:    20,
			30:    30,
			45:    60, // Round 45s to 60s
			60:    60,
			90:    60, // Round 90s to 60s
			120:   120,
			180:   180,
			240:   300, // Round 4min to 5min
			300:   300,
			600:   600,
			900:   600, // Round 15min to 10min
			1800:  1800,
			3600:  3600,
			7200:  3600, // Round 2hr to 1hr
			21600: 21600,
			43200: 43200,
			86400: 86400,
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

	for _, m := range monitors {
		cm, monitorIssues := c.convertMonitor(m)
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
	regions := c.mapRegions(attrs.Regions)
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

	for _, h := range heartbeats {
		ch, heartbeatIssues := c.convertHeartbeat(h)
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

func (c *Converter) mapFrequency(frequency int) int {
	if mapped, ok := c.frequencyMap[frequency]; ok {
		return mapped
	}

	// Find closest supported frequency
	supported := []int{10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400}
	closest := supported[0]
	minDiff := abs(frequency - closest)

	for _, f := range supported {
		diff := abs(frequency - f)
		if diff < minDiff {
			minDiff = diff
			closest = f
		}
	}

	return closest
}

func (c *Converter) mapRegions(bsRegions []string) []string {
	var regions []string
	seen := make(map[string]bool)

	for _, region := range bsRegions {
		normalized := strings.ToLower(strings.TrimSpace(region))
		if mapped, ok := c.regionMap[normalized]; ok {
			if !seen[mapped] {
				regions = append(regions, mapped)
				seen[mapped] = true
			}
		}
	}

	return regions
}

func sanitizeResourceName(name string) string {
	// Convert to lowercase
	safe := strings.ToLower(name)

	// Replace special characters with underscores
	reg := regexp.MustCompile(`[^a-z0-9_]+`)
	safe = reg.ReplaceAllString(safe, "_")

	// Remove leading/trailing underscores
	safe = strings.Trim(safe, "_")

	// Collapse multiple underscores
	reg = regexp.MustCompile(`_+`)
	safe = reg.ReplaceAllString(safe, "_")

	// Ensure it doesn't start with a digit
	if safe != "" && safe[0] >= '0' && safe[0] <= '9' {
		safe = "monitor_" + safe
	}

	// Ensure it's not empty
	if safe == "" {
		safe = "unnamed_monitor"
	}

	return safe
}

func extractIssueMessages(issues []ConversionIssue) []string {
	var messages []string
	for _, issue := range issues {
		messages = append(messages, issue.Message)
	}
	return messages
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

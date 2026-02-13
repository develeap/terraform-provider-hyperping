// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"fmt"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// ConversionResult represents the result of converting a Pingdom check.
type ConversionResult struct {
	Monitor         *client.CreateMonitorRequest
	Healthcheck     *client.CreateHealthcheckRequest
	Supported       bool
	UnsupportedType string
	Notes           []string
}

// CheckConverter converts Pingdom checks to Hyperping resources.
type CheckConverter struct{}

// NewCheckConverter creates a new CheckConverter.
func NewCheckConverter() *CheckConverter {
	return &CheckConverter{}
}

// Convert converts a Pingdom check to a Hyperping resource.
func (c *CheckConverter) Convert(check pingdom.Check) ConversionResult {
	result := ConversionResult{
		Notes: []string{},
	}

	switch check.Type {
	case "http", "https":
		result.Monitor = c.convertHTTPCheck(check)
		result.Supported = true
	case "tcp":
		result.Monitor = c.convertTCPCheck(check)
		result.Supported = true
	case "ping":
		result.Monitor = c.convertPingCheck(check)
		result.Supported = true
	case "smtp":
		result.Monitor = c.convertSMTPCheck(check)
		result.Supported = true
		result.Notes = append(result.Notes, "SMTP check converted to TCP port check on port 25/587")
	case "pop3":
		result.Monitor = c.convertPOP3Check(check)
		result.Supported = true
		result.Notes = append(result.Notes, "POP3 check converted to TCP port check on port 110/995")
	case "imap":
		result.Monitor = c.convertIMAPCheck(check)
		result.Supported = true
		result.Notes = append(result.Notes, "IMAP check converted to TCP port check on port 143/993")
	case "dns":
		result.Supported = false
		result.UnsupportedType = "dns"
		result.Notes = append(result.Notes, "DNS checks not directly supported. Consider using HTTP check to DNS-over-HTTPS service or monitor the service that relies on DNS")
	case "udp":
		result.Supported = false
		result.UnsupportedType = "udp"
		result.Notes = append(result.Notes, "UDP checks not supported. Consider using TCP alternative if available")
	case "transaction":
		result.Supported = false
		result.UnsupportedType = "transaction"
		result.Notes = append(result.Notes, "Transaction checks not directly supported. Break into individual HTTP monitors or use external script with healthcheck")
	default:
		result.Supported = false
		result.UnsupportedType = check.Type
		result.Notes = append(result.Notes, fmt.Sprintf("Unknown check type: %s", check.Type))
	}

	return result
}

func (c *CheckConverter) convertHTTPCheck(check pingdom.Check) *client.CreateMonitorRequest {
	// Build URL
	protocol := "http"
	if check.Encryption {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s%s", protocol, check.Hostname, check.URL)

	// Convert frequency (minutes to seconds)
	frequency := ConvertFrequency(check.Resolution)

	// Build request headers
	headers := make([]client.RequestHeader, 0, len(check.RequestHeaders))
	for name, value := range check.RequestHeaders {
		headers = append(headers, client.RequestHeader{
			Name:  name,
			Value: value,
		})
	}

	// Convert regions
	regions := ConvertRegions(check.ProbeFilters)

	monitor := &client.CreateMonitorRequest{
		Name:            GenerateName(check),
		URL:             url,
		Protocol:        "http",
		HTTPMethod:      "GET",
		CheckFrequency:  frequency,
		Regions:         regions,
		RequestHeaders:  headers,
		FollowRedirects: boolPtr(true),
		Paused:          check.Paused,
	}

	// Handle POST data
	if check.PostData != "" {
		monitor.HTTPMethod = "POST"
		monitor.RequestBody = &check.PostData
	}

	// Handle expected status
	monitor.ExpectedStatusCode = "200"

	// Handle body content matching
	if check.ShouldContain != "" {
		monitor.RequiredKeyword = &check.ShouldContain
	}

	// Handle SSL verification
	if check.Encryption {
		followRedirects := !check.VerifyCertificate
		monitor.FollowRedirects = &followRedirects
	}

	return monitor
}

func (c *CheckConverter) convertTCPCheck(check pingdom.Check) *client.CreateMonitorRequest {
	frequency := ConvertFrequency(check.Resolution)
	regions := ConvertRegions(check.ProbeFilters)

	port := check.Port
	if port == 0 {
		port = 80
	}

	return &client.CreateMonitorRequest{
		Name:           GenerateName(check),
		URL:            check.Hostname,
		Protocol:       "port",
		CheckFrequency: frequency,
		Regions:        regions,
		Port:           &port,
		Paused:         check.Paused,
	}
}

func (c *CheckConverter) convertPingCheck(check pingdom.Check) *client.CreateMonitorRequest {
	frequency := ConvertFrequency(check.Resolution)
	regions := ConvertRegions(check.ProbeFilters)

	return &client.CreateMonitorRequest{
		Name:           GenerateName(check),
		URL:            check.Hostname,
		Protocol:       "icmp",
		CheckFrequency: frequency,
		Regions:        regions,
		Paused:         check.Paused,
	}
}

func (c *CheckConverter) convertSMTPCheck(check pingdom.Check) *client.CreateMonitorRequest {
	frequency := ConvertFrequency(check.Resolution)
	regions := ConvertRegions(check.ProbeFilters)

	port := check.Port
	if port == 0 {
		port = 25
		if check.Encryption {
			port = 587
		}
	}

	return &client.CreateMonitorRequest{
		Name:           GenerateName(check),
		URL:            check.Hostname,
		Protocol:       "port",
		CheckFrequency: frequency,
		Regions:        regions,
		Port:           &port,
		Paused:         check.Paused,
	}
}

func (c *CheckConverter) convertPOP3Check(check pingdom.Check) *client.CreateMonitorRequest {
	frequency := ConvertFrequency(check.Resolution)
	regions := ConvertRegions(check.ProbeFilters)

	port := check.Port
	if port == 0 {
		port = 110
		if check.Encryption {
			port = 995
		}
	}

	return &client.CreateMonitorRequest{
		Name:           GenerateName(check),
		URL:            check.Hostname,
		Protocol:       "port",
		CheckFrequency: frequency,
		Regions:        regions,
		Port:           &port,
		Paused:         check.Paused,
	}
}

func (c *CheckConverter) convertIMAPCheck(check pingdom.Check) *client.CreateMonitorRequest {
	frequency := ConvertFrequency(check.Resolution)
	regions := ConvertRegions(check.ProbeFilters)

	port := check.Port
	if port == 0 {
		port = 143
		if check.Encryption {
			port = 993
		}
	}

	return &client.CreateMonitorRequest{
		Name:           GenerateName(check),
		URL:            check.Hostname,
		Protocol:       "port",
		CheckFrequency: frequency,
		Regions:        regions,
		Port:           &port,
		Paused:         check.Paused,
	}
}

// ConvertFrequency converts Pingdom resolution (minutes) to Hyperping frequency (seconds).
func ConvertFrequency(resolutionMinutes int) int {
	seconds := resolutionMinutes * 60

	// Round to nearest allowed frequency
	allowed := []int{60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400}

	closest := allowed[0]
	minDiff := abs(seconds - allowed[0])

	for _, freq := range allowed {
		diff := abs(seconds - freq)
		if diff < minDiff {
			minDiff = diff
			closest = freq
		}
	}

	return closest
}

// ConvertRegions converts Pingdom probe filters to Hyperping regions.
func ConvertRegions(probeFilters []string) []string {
	if len(probeFilters) == 0 {
		// Default regions
		return []string{"virginia", "london", "frankfurt", "singapore"}
	}

	regionMap := map[string][]string{
		"region:NA":    {"virginia", "oregon"},
		"region:EU":    {"london", "frankfurt"},
		"region:APAC":  {"singapore", "sydney", "tokyo"},
		"region:LATAM": {"saopaulo"},
	}

	regionsSet := make(map[string]bool)
	for _, filter := range probeFilters {
		if regions, ok := regionMap[filter]; ok {
			for _, r := range regions {
				regionsSet[r] = true
			}
		}
	}

	if len(regionsSet) == 0 {
		return []string{"virginia", "london"}
	}

	regions := make([]string, 0, len(regionsSet))
	for r := range regionsSet {
		regions = append(regions, r)
	}

	return regions
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func boolPtr(b bool) *bool {
	return &b
}

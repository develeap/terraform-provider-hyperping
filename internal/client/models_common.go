// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// =============================================================================
// Common Types
// =============================================================================

// FlexibleString is a string type that can unmarshal from both JSON strings and numbers.
// This handles API inconsistencies where a field might be returned as either type.
type FlexibleString string

// maxFlexibleStringBytes is the maximum allowed input size for FlexibleString.
// Prevents memory exhaustion from malicious or malformed numeric strings (VULN-004).
const maxFlexibleStringBytes = 100

// UnmarshalJSON implements json.Unmarshaler for FlexibleString.
func (fs *FlexibleString) UnmarshalJSON(data []byte) error {
	if len(data) > maxFlexibleStringBytes {
		return fmt.Errorf("FlexibleString input exceeds maximum size of %d bytes", maxFlexibleStringBytes)
	}

	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fs = FlexibleString(s)
		return nil
	}

	// Try number
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*fs = FlexibleString(n.String())
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into FlexibleString", string(data))
}

// MarshalJSON implements json.Marshaler for FlexibleString.
func (fs FlexibleString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(fs))
}

// String returns the string value.
func (fs FlexibleString) String() string {
	return string(fs)
}

// Input length limits to prevent resource exhaustion (VULN-007).
const (
	maxNameLength    = 255
	maxURLLength     = 2048
	maxMessageLength = 10000
)

// validateStringLength checks that a string does not exceed the given max length
// measured in Unicode code points (runes), not bytes (VULN-018).
// Multi-byte characters (e.g. CJK, emoji) count as one character each.
func validateStringLength(field, value string, maxLen int) error {
	runeCount := utf8.RuneCountInString(value)
	if runeCount > maxLen {
		return fmt.Errorf("field %q exceeds maximum length of %d characters (got %d)", field, maxLen, runeCount)
	}
	return nil
}

// localizedField pairs a language code with its value for deterministic iteration.
type localizedField struct {
	lang string
	val  string
}

// localizedFields returns all locale fields in a stable, deterministic order.
func localizedFields(text LocalizedText) []localizedField {
	return []localizedField{
		{"en", text.En}, {"fr", text.Fr}, {"de", text.De}, {"ru", text.Ru},
		{"nl", text.Nl}, {"es", text.Es}, {"it", text.It}, {"pt", text.Pt},
		{"ja", text.Ja}, {"zh", text.Zh},
	}
}

// validateLocalizedText validates all non-empty locale fields of a LocalizedText value.
// Iteration order is deterministic (en, fr, de, ru, nl, es, it, pt, ja, zh).
func validateLocalizedText(prefix string, text LocalizedText, maxLen int) error {
	for _, f := range localizedFields(text) {
		if f.val != "" {
			if err := validateStringLength(prefix+"."+f.lang, f.val, maxLen); err != nil {
				return err
			}
		}
	}
	return nil
}

// RequestHeader represents a single HTTP header for monitor requests.
// API format: [{"name": "Header-Name", "value": "header-value"}]
type RequestHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// LocalizedText represents text that can be localized in multiple languages.
// Used for incident/maintenance titles and descriptions.
// API format: {"en": "English text", "fr": "French text"}
// Supports all AllowedLanguages: en, fr, de, ru, nl, es, it, pt, ja, zh.
type LocalizedText struct {
	En string `json:"en,omitempty"`
	Fr string `json:"fr,omitempty"`
	De string `json:"de,omitempty"`
	Ru string `json:"ru,omitempty"`
	Nl string `json:"nl,omitempty"`
	Es string `json:"es,omitempty"`
	It string `json:"it,omitempty"`
	Pt string `json:"pt,omitempty"`
	Ja string `json:"ja,omitempty"`
	Zh string `json:"zh,omitempty"`
}

// =============================================================================
// Allowed Values
// =============================================================================

const (
	// DefaultMonitorFrequency is the default check frequency for monitors in seconds.
	// This is used when no frequency is specified in the API request.
	DefaultMonitorFrequency = 60

	// DefaultMonitorTimeout is the default timeout for monitor checks in seconds.
	// This is used when no timeout is specified in the API request.
	DefaultMonitorTimeout = 10

	// DefaultNotifyBeforeMinutes is the default number of minutes before maintenance
	// to notify subscribers. This is used when no notification time is specified.
	DefaultNotifyBeforeMinutes = 60
)

var (
	// AllowedFrequencies contains valid monitor check frequencies in seconds.
	AllowedFrequencies = []int{10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400}

	// AllowedTimeouts contains valid monitor timeout values in seconds.
	AllowedTimeouts = []int{5, 10, 15, 20}

	// AllowedProtocols contains valid monitor protocols.
	AllowedProtocols = []string{"http", "port", "icmp", "dns"}

	// AllowedDNSRecordTypes contains valid DNS record types for DNS-protocol monitors.
	AllowedDNSRecordTypes = []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT", "SOA", "SRV", "CAA", "PTR"}

	// AllowedMethods contains valid HTTP methods for monitors.
	AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	// AllowedRegions contains valid monitor check regions.
	// Combined from official Hyperping API documentation and real API responses.
	// See: https://hyperping.com/docs/api/monitors/create
	AllowedRegions = []string{
		// Europe
		"london",
		"frankfurt",
		"paris",
		"amsterdam",
		// Asia Pacific
		"singapore",
		"sydney",
		"tokyo",
		"seoul",
		"mumbai",
		"bangalore",
		// North America
		"virginia",
		"california",
		"sanfrancisco",
		"nyc",
		"toronto",
		// South America
		"saopaulo",
		// Middle East
		"bahrain",
		// Africa
		"capetown",
	}

	// AllowedIncidentTypes contains valid incident type values.
	AllowedIncidentTypes = []string{"outage", "incident"}

	// AllowedIncidentUpdateTypes contains valid incident update type values.
	AllowedIncidentUpdateTypes = []string{"investigating", "identified", "update", "monitoring", "resolved"}

	// AllowedNotificationOptions contains valid maintenance notification options.
	// "none" disables notifications, "scheduled" sends before start, "immediate" sends at creation.
	AllowedNotificationOptions = []string{"none", "scheduled", "immediate"}

	// AllowedPeriodTypes contains valid healthcheck period type values.
	AllowedPeriodTypes = []string{"seconds", "minutes", "hours", "days"}

	// AllowedStatusPageThemes contains valid status page theme values.
	AllowedStatusPageThemes = []string{"light", "dark", "system"}

	// AllowedStatusPageFonts contains valid status page font values.
	AllowedStatusPageFonts = []string{
		"system-ui", "Lato", "Manrope", "Inter", "Open Sans",
		"Montserrat", "Poppins", "Roboto", "Raleway", "Nunito",
		"Merriweather", "DM Sans", "Work Sans",
	}

	// AllowedLanguages contains valid language codes for status page configuration.
	// From API spec: https://hyperping.com/docs/api/status-pages/create
	// Note: LocalizedText struct supports additional languages for content fields,
	// but the status page `languages` setting only accepts these values.
	AllowedLanguages = []string{"en", "fr", "de", "ru", "nl", "pl", "sv"}

	// AllowedSubscriberTypes contains valid subscriber type values.
	AllowedSubscriberTypes = []string{"email", "sms", "teams"}
)

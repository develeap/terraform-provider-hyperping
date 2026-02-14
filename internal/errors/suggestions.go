// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package terraflyerrors

import (
	"fmt"
	"sort"
	"strings"
)

// ValidationSuggestion generates suggestions for validation errors with allowed values.
func ValidationSuggestion(field, currentValue string, allowedValues []interface{}) *EnhancedError {
	allowedStr := formatAllowedValues(allowedValues)

	description := fmt.Sprintf("The '%s' field has an invalid value.\n\nCurrent value: %s\nAllowed values: %s",
		field, currentValue, allowedStr)

	suggestions := []string{
		fmt.Sprintf("Change the '%s' field to one of the allowed values", field),
	}

	// Try to find closest value
	if closest := findClosestValue(currentValue, allowedValues); closest != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("Did you mean '%s'? (closest match)", closest))
	}

	examples := []string{}
	for _, v := range allowedValues {
		examples = append(examples, fmt.Sprintf("%s = %v", field, v))
		if len(examples) >= 3 {
			break
		}
	}

	return &EnhancedError{
		Title:       "Invalid Field Value",
		Description: description,
		Field:       field,
		Suggestions: suggestions,
		Examples:    examples,
	}
}

// FrequencySuggestion generates suggestions for monitor frequency validation errors.
func FrequencySuggestion(currentValue int64) *EnhancedError {
	allowedFrequencies := []int64{10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400}

	// Find closest frequencies
	closest := findClosestFrequencies(currentValue, allowedFrequencies, 2)

	description := fmt.Sprintf("The 'check_frequency' field must be one of the allowed values.\n\n"+
		"Current value: %d seconds (invalid)\n"+
		"Allowed values (in seconds): 10, 20, 30, 60, 120, 180, 300, 600, 1800, 3600, 21600, 43200, 86400",
		currentValue)

	suggestions := make([]string, 0, len(closest))
	for _, freq := range closest {
		diff := freq - currentValue
		var diffStr string
		if diff > 0 {
			diffStr = fmt.Sprintf("%d seconds slower", diff)
		} else {
			diffStr = fmt.Sprintf("%d seconds faster", -diff)
		}
		suggestions = append(suggestions,
			fmt.Sprintf("Use %d seconds (%s)", freq, diffStr))
	}

	examples := []string{
		"check_frequency = 30   # Check every 30 seconds",
		"check_frequency = 60   # Check every minute",
		"check_frequency = 300  # Check every 5 minutes",
	}

	return &EnhancedError{
		Title:       "Invalid Monitor Frequency",
		Description: description,
		Field:       "check_frequency",
		Suggestions: suggestions,
		Examples:    examples,
		DocLinks: []string{
			"https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/monitor#check_frequency",
		},
	}
}

// RegionSuggestion generates suggestions for invalid region values.
func RegionSuggestion(invalidRegion string) *EnhancedError {
	allowedRegions := []string{
		"london", "frankfurt", "singapore", "sydney",
		"virginia", "oregon", "saopaulo", "tokyo", "bahrain",
	}

	description := fmt.Sprintf("The region '%s' is not valid.\n\n"+
		"Allowed regions: %s",
		invalidRegion, strings.Join(allowedRegions, ", "))

	// Try to find closest match
	suggestions := make([]string, 0, 1)
	if closest := findClosestString(invalidRegion, allowedRegions); closest != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("Did you mean '%s'?", closest))
	}

	examples := []string{
		`regions = ["london", "frankfurt"]          # Europe`,
		`regions = ["virginia", "oregon"]           # North America`,
		`regions = ["singapore", "sydney", "tokyo"] # Asia-Pacific`,
	}

	return &EnhancedError{
		Title:       "Invalid Region",
		Description: description,
		Field:       "regions",
		Suggestions: suggestions,
		Examples:    examples,
		DocLinks: []string{
			"https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/monitor#regions",
		},
	}
}

// HTTPMethodSuggestion generates suggestions for invalid HTTP method values.
func HTTPMethodSuggestion(invalidMethod string) *EnhancedError {
	allowedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}

	description := fmt.Sprintf("The HTTP method '%s' is not valid.\n\n"+
		"Allowed methods: %s",
		invalidMethod, strings.Join(allowedMethods, ", "))

	// Try to find closest match (case-insensitive)
	suggestions := []string{}
	normalized := strings.ToUpper(invalidMethod)
	for _, method := range allowedMethods {
		if method == normalized {
			suggestions = append(suggestions,
				fmt.Sprintf("Did you mean '%s'? (check capitalization)", method))
			break
		}
	}

	if len(suggestions) == 0 {
		if closest := findClosestString(strings.ToLower(invalidMethod), lowerStrings(allowedMethods)); closest != "" {
			suggestions = append(suggestions,
				fmt.Sprintf("Did you mean '%s'?", strings.ToUpper(closest)))
		}
	}

	examples := []string{
		`http_method = "GET"   # Most common for health checks`,
		`http_method = "POST"  # For endpoints requiring POST`,
		`http_method = "HEAD"  # For lightweight checks`,
	}

	return &EnhancedError{
		Title:       "Invalid HTTP Method",
		Description: description,
		Field:       "http_method",
		Suggestions: suggestions,
		Examples:    examples,
		DocLinks: []string{
			"https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/monitor#http_method",
		},
	}
}

// StatusCodeSuggestion generates suggestions for expected status code validation errors.
func StatusCodeSuggestion(currentValue string) *EnhancedError {
	description := fmt.Sprintf("The 'expected_status_code' field must be a valid HTTP status code or range.\n\n"+
		"Current value: %s\n"+
		"Valid formats:\n"+
		"  • Single code: \"200\", \"404\", \"500\"\n"+
		"  • Range: \"2xx\", \"4xx\", \"5xx\"\n"+
		"  • Multiple: \"200,201,202\"",
		currentValue)

	examples := []string{
		`expected_status_code = "200"       # Expect OK`,
		`expected_status_code = "2xx"       # Any success code`,
		`expected_status_code = "200,201"   # Multiple codes`,
	}

	return &EnhancedError{
		Title:       "Invalid Status Code",
		Description: description,
		Field:       "expected_status_code",
		Examples:    examples,
		DocLinks: []string{
			"https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/monitor#expected_status_code",
		},
	}
}

// IncidentStatusSuggestion generates suggestions for incident status validation errors.
func IncidentStatusSuggestion(invalidStatus string) *EnhancedError {
	allowedStatuses := []string{"investigating", "identified", "monitoring", "resolved"}

	description := fmt.Sprintf("The incident status '%s' is not valid.\n\n"+
		"Allowed statuses: %s\n\n"+
		"Status workflow: investigating → identified → monitoring → resolved",
		invalidStatus, strings.Join(allowedStatuses, ", "))

	suggestions := []string{}
	if closest := findClosestString(invalidStatus, allowedStatuses); closest != "" {
		suggestions = append(suggestions,
			fmt.Sprintf("Did you mean '%s'?", closest))
	}

	examples := []string{
		`status = "investigating"  # Initial status`,
		`status = "identified"     # Root cause found`,
		`status = "monitoring"     # Fix deployed, monitoring`,
		`status = "resolved"       # Incident resolved`,
	}

	return &EnhancedError{
		Title:       "Invalid Incident Status",
		Description: description,
		Field:       "status",
		Suggestions: suggestions,
		Examples:    examples,
		DocLinks: []string{
			"https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/incident#status",
		},
	}
}

// SeveritySuggestion generates suggestions for incident severity validation errors.
func SeveritySuggestion(invalidSeverity string) *EnhancedError {
	allowedSeverities := []string{"minor", "major", "critical"}

	description := fmt.Sprintf("The incident severity '%s' is not valid.\n\n"+
		"Allowed severities: %s",
		invalidSeverity, strings.Join(allowedSeverities, ", "))

	suggestions := []string{
		"minor - Low impact, degraded performance",
		"major - Significant impact, partial outage",
		"critical - Severe impact, complete outage",
	}

	if closest := findClosestString(invalidSeverity, allowedSeverities); closest != "" {
		suggestions = append([]string{fmt.Sprintf("Did you mean '%s'?", closest)}, suggestions...)
	}

	examples := []string{
		`severity = "minor"     # Slow response times`,
		`severity = "major"     # API partially unavailable`,
		`severity = "critical"  # Complete outage`,
	}

	return &EnhancedError{
		Title:       "Invalid Incident Severity",
		Description: description,
		Field:       "severity",
		Suggestions: suggestions,
		Examples:    examples,
		DocLinks: []string{
			"https://registry.terraform.io/providers/develeap/hyperping/latest/docs/resources/incident#severity",
		},
	}
}

// formatAllowedValues formats a slice of allowed values for display.
func formatAllowedValues(values []interface{}) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, fmt.Sprintf("%v", v))
	}
	return strings.Join(parts, ", ")
}

// findClosestValue finds the closest matching value from allowed values.
func findClosestValue(current string, allowed []interface{}) string {
	candidates := make([]string, 0, len(allowed))
	for _, v := range allowed {
		candidates = append(candidates, fmt.Sprintf("%v", v))
	}
	return findClosestString(current, candidates)
}

// findClosestString finds the closest string match using simple heuristics.
func findClosestString(target string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}

	target = strings.ToLower(target)
	minDist := len(target) + 100
	closest := ""
	isPrefixMatch := false

	for _, candidate := range candidates {
		cand := strings.ToLower(candidate)

		// Exact match
		if target == cand {
			return candidate
		}

		// Prefix match - always accept prefix matches regardless of length difference
		if strings.HasPrefix(cand, target) || strings.HasPrefix(target, cand) {
			dist := abs(len(target) - len(cand))
			if dist < minDist {
				minDist = dist
				closest = candidate
				isPrefixMatch = true
			}
			continue
		}

		// Levenshtein distance (only if we haven't found a prefix match)
		if !isPrefixMatch {
			dist := levenshtein(target, cand)
			if dist < minDist {
				minDist = dist
				closest = candidate
			}
		}
	}

	// Accept prefix matches always, otherwise only suggest if reasonably close
	if isPrefixMatch || minDist <= len(target)/2 {
		return closest
	}

	return ""
}

// findClosestFrequencies finds the N closest frequency values.
func findClosestFrequencies(target int64, candidates []int64, n int) []int64 {
	if len(candidates) == 0 {
		return nil
	}

	type distPair struct {
		value int64
		dist  int64
	}

	var pairs []distPair
	for _, cand := range candidates {
		dist := cand - target
		if dist < 0 {
			dist = -dist
		}
		pairs = append(pairs, distPair{value: cand, dist: dist})
	}

	// Sort by distance
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].dist < pairs[j].dist
	})

	// Return top N
	result := []int64{}
	for i := 0; i < n && i < len(pairs); i++ {
		result = append(result, pairs[i].value)
	}

	return result
}

// levenshtein calculates the Levenshtein distance between two strings.
func levenshtein(s1, s2 string) int {
	if s1 == "" {
		return len(s2)
	}
	if s2 == "" {
		return len(s1)
	}

	// Create distance matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Calculate distances
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = minInt(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func minInt(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func lowerStrings(strs []string) []string {
	result := make([]string, len(strs))
	for i, s := range strs {
		result[i] = strings.ToLower(s)
	}
	return result
}

// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package converter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

// GenerateName generates a Hyperping-style name from a Pingdom check.
// Format: [ENVIRONMENT]-Category-ServiceName
func GenerateName(check pingdom.Check) string {
	environment := extractEnvironment(check.Tags)
	category := extractCategory(check.Tags)
	customer := extractCustomer(check.Tags)
	serviceName := sanitizeServiceName(check.Name)

	if customer != "" {
		return fmt.Sprintf("[%s-%s]-%s-%s", environment, customer, category, serviceName)
	}

	return fmt.Sprintf("[%s]-%s-%s", environment, category, serviceName)
}

// extractEnvironment extracts environment from tags.
func extractEnvironment(tags []pingdom.Tag) string {
	envMap := map[string]string{
		"production":  "PROD",
		"prod":        "PROD",
		"staging":     "STAGING",
		"stage":       "STAGING",
		"development": "DEV",
		"dev":         "DEV",
		"qa":          "QA",
		"test":        "TEST",
	}

	for _, tag := range tags {
		tagName := strings.ToLower(tag.Name)
		if env, ok := envMap[tagName]; ok {
			return env
		}
	}

	return "UNKNOWN"
}

// extractCategory extracts service category from tags.
func extractCategory(tags []pingdom.Tag) string {
	categories := map[string]string{
		"api":         "API",
		"web":         "Web",
		"website":     "Web",
		"database":    "Database",
		"db":          "Database",
		"cache":       "Cache",
		"redis":       "Cache",
		"memcached":   "Cache",
		"queue":       "Queue",
		"worker":      "Worker",
		"cdn":         "CDN",
		"dns":         "DNS",
		"mail":        "Mail",
		"smtp":        "Mail",
		"email":       "Mail",
		"frontend":    "Frontend",
		"backend":     "Backend",
		"service":     "Service",
		"app":         "App",
		"application": "App",
	}

	for _, tag := range tags {
		tagName := strings.ToLower(tag.Name)
		if category, ok := categories[tagName]; ok {
			return category
		}
	}

	return "Service"
}

// extractCustomer extracts customer/tenant from tags.
func extractCustomer(tags []pingdom.Tag) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag.Name, "customer-") {
			customer := strings.TrimPrefix(tag.Name, "customer-")
			return strings.ToUpper(customer)
		}
		if strings.HasPrefix(tag.Name, "tenant-") {
			tenant := strings.TrimPrefix(tag.Name, "tenant-")
			return strings.ToUpper(tenant)
		}
	}

	return ""
}

// sanitizeServiceName cleans up service name for use in Hyperping naming.
func sanitizeServiceName(name string) string {
	// Remove common prefixes
	prefixPattern := regexp.MustCompile(`(?i)^(Production|Staging|Dev|QA|Test)\s*-?\s*`)
	name = prefixPattern.ReplaceAllString(name, "")

	// Remove special characters except spaces and hyphens
	specialChars := regexp.MustCompile(`[^\w\s-]`)
	name = specialChars.ReplaceAllString(name, "")

	// Convert to title case and remove spaces
	words := strings.Fields(name)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	name = strings.Join(words, "")

	// Limit length
	if len(name) > 30 {
		name = name[:30]
	}

	// Fallback for empty names
	if name == "" {
		name = "Monitor"
	}

	return name
}

// TagsToString converts tags to a comma-separated string for display.
func TagsToString(tags []pingdom.Tag) string {
	if len(tags) == 0 {
		return ""
	}

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return strings.Join(tagNames, ", ")
}

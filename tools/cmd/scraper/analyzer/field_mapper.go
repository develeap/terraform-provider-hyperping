package analyzer

import (
	"regexp"
	"strings"
)

// ResourceMapping associates API endpoints with Terraform resources
type ResourceMapping struct {
	TerraformResource   string            // e.g., "hyperping_monitor"
	APISection          string            // e.g., "monitors"
	APIEndpoints        []string          // e.g., ["monitors_create", "monitors_update"]
	FieldOverrides      map[string]string // API field -> TF field for special cases
	NestedFieldMappings map[string]string // API field -> nested TF path (e.g., "website" -> "settings.website")
}

// SkippedFields contains fields that are intentionally not exposed in Terraform
var SkippedFields = map[string]bool{
	// Read-only timestamps
	"status":         true,
	"createdAt":      true,
	"created_at":     true,
	"updatedAt":      true,
	"updated_at":     true,

	// Internal IDs (exposed as "id")
	"uuid":           true,
	"projectUuid":    true,
	"project_uuid":   true,

	// Computed fields
	"sslExpiration":  true,
	"ssl_expiration": true,

	// Internal references
	"bulkUuid":       true,
	"bulk_uuid":      true,
	"createdBy":      true,
	"created_by":     true,

	// Separate resources
	"updates":        true, // Nested updates array (hyperping_incident_update)
	"subscriberId":   true, // Belongs to hyperping_statuspage_subscriber
	"subscriber_id":  true,

	// Query/pagination parameters (not resource fields)
	"page":           true,
	"search":         true,
	"limit":          true,
	"offset":         true,
	"type":           true, // Filter param for list endpoints (e.g., outages?type=manual)

	// Users array (complex nested, handled separately)
	"users":          true,
}

// ResourceMappings defines all known API-to-Terraform resource mappings
var ResourceMappings = []ResourceMapping{
	{
		TerraformResource: "hyperping_monitor",
		APISection:        "monitors",
		APIEndpoints:      []string{"monitors", "monitors_create", "monitors_update", "monitors_delete", "monitors_get", "monitors_list"},
		FieldOverrides: map[string]string{
			"uuid":         "id",
			"monitorUuid":  "id",
			"monitor_uuid": "id",
		},
	},
	{
		TerraformResource: "hyperping_incident",
		APISection:        "incidents",
		APIEndpoints:      []string{"incidents", "incidents_create", "incidents_update", "incidents_delete", "incidents_get", "incidents_list"},
		FieldOverrides: map[string]string{
			"uuid":         "id",
			"statuspages":  "status_pages",
		},
	},
	{
		TerraformResource: "hyperping_maintenance",
		APISection:        "maintenance",
		APIEndpoints:      []string{"maintenance", "maintenance_create", "maintenance_update", "maintenance_delete", "maintenance_get", "maintenance_list"},
		FieldOverrides: map[string]string{
			"uuid":        "id",
			"statuspages": "status_pages",
		},
	},
	{
		TerraformResource: "hyperping_healthcheck",
		APISection:        "healthchecks",
		APIEndpoints:      []string{"healthchecks", "healthchecks_create", "healthchecks_update", "healthchecks_delete", "healthchecks_get", "healthchecks_list", "healthchecks_pause", "healthchecks_resume"},
		FieldOverrides: map[string]string{
			"uuid":             "id",
			"healthcheckUuid":  "id",
			"healthcheck_uuid": "id",
			"pingUrl":          "ping_url",
			"tz":               "timezone", // API accepts both tz and timezone
		},
	},
	{
		TerraformResource: "hyperping_statuspage",
		APISection:        "statuspages",
		APIEndpoints:      []string{"statuspages", "statuspages_create", "statuspages_update", "statuspages_delete", "statuspages_get", "statuspages_list"},
		FieldOverrides: map[string]string{
			"uuid":      "id",
			"subdomain": "hosted_subdomain", // API uses subdomain, TF uses hosted_subdomain
		},
		// StatusPage API has flat structure, but TF nests many fields inside "settings" for better UX
		NestedFieldMappings: map[string]string{
			"website":                 "settings.website",
			"description":             "settings.description",
			"languages":               "settings.languages",
			"theme":                   "settings.theme",
			"font":                    "settings.font",
			"accent_color":            "settings.accent_color",
			"auto_refresh":            "settings.auto_refresh",
			"banner_header":           "settings.banner_header",
			"logo":                    "settings.logo",
			"logo_height":             "settings.logo_height",
			"favicon":                 "settings.favicon",
			"hide_powered_by":         "settings.hide_powered_by",
			"hide_from_search_engines": "settings.hide_from_search_engines",
			"google_analytics":        "settings.google_analytics",
			"subscribe":               "settings.subscribe",
			"authentication":          "settings.authentication",
		},
	},
	{
		TerraformResource: "hyperping_outage",
		APISection:        "outages",
		APIEndpoints:      []string{"outages", "outages_create", "outages_acknowledge", "outages_unacknowledge", "outages_resolve", "outages_escalate", "outages_delete", "outages_get", "outages_list"},
		FieldOverrides: map[string]string{
			"uuid":        "id",
			"outageUuid":  "id",
			"outage_uuid": "id",
		},
	},
	{
		TerraformResource: "hyperping_statuspage_subscriber",
		APISection:        "statuspages_subscribers",
		APIEndpoints:      []string{"statuspages_subscribers", "statuspages_subscribers_create", "statuspages_subscribers_delete", "statuspages_subscribers_list"},
		FieldOverrides: map[string]string{
			"uuid":           "id",
			"subscriberUuid": "id",
			"subscriber_uuid": "id",
		},
	},
}

// GetResourceMapping returns the mapping for an API section
func GetResourceMapping(apiSection string) *ResourceMapping {
	// Normalize section name
	section := strings.ToLower(apiSection)

	for i := range ResourceMappings {
		if ResourceMappings[i].APISection == section {
			return &ResourceMappings[i]
		}
	}
	return nil
}

// GetMappingByEndpoint returns the mapping for a specific API endpoint file
func GetMappingByEndpoint(endpointFile string) *ResourceMapping {
	// endpointFile is like "monitors_create.json" -> extract "monitors"
	parts := strings.Split(strings.TrimSuffix(endpointFile, ".json"), "_")
	if len(parts) == 0 {
		return nil
	}

	section := parts[0]
	return GetResourceMapping(section)
}

// MapAPIFieldToTerraform converts an API field name to its Terraform equivalent
func MapAPIFieldToTerraform(apiField string, mapping *ResourceMapping) string {
	// Check for explicit overrides first
	if mapping != nil && mapping.FieldOverrides != nil {
		if tfField, ok := mapping.FieldOverrides[apiField]; ok {
			return tfField
		}
	}

	// Convert camelCase to snake_case
	return CamelToSnake(apiField)
}

// GetNestedFieldMapping returns the nested TF path for an API field, if it exists
// Returns empty string if the field is not mapped to a nested path
func GetNestedFieldMapping(apiField string, mapping *ResourceMapping) string {
	if mapping == nil || mapping.NestedFieldMappings == nil {
		return ""
	}
	// Check both original name and snake_case version
	if path, ok := mapping.NestedFieldMappings[apiField]; ok {
		return path
	}
	snakeName := CamelToSnake(apiField)
	if path, ok := mapping.NestedFieldMappings[snakeName]; ok {
		return path
	}
	return ""
}

// IsNestedField checks if an API field maps to a nested TF field
func IsNestedField(apiField string, mapping *ResourceMapping) bool {
	return GetNestedFieldMapping(apiField, mapping) != ""
}

// MapTerraformFieldToAPI converts a Terraform field name to its API equivalent
func MapTerraformFieldToAPI(tfField string, mapping *ResourceMapping) string {
	// Check for explicit overrides (reverse lookup)
	if mapping != nil && mapping.FieldOverrides != nil {
		for apiField, mappedTF := range mapping.FieldOverrides {
			if mappedTF == tfField {
				return apiField
			}
		}
	}

	// For most fields, they're already in snake_case in both
	return tfField
}

// CamelToSnake converts camelCase to snake_case
func CamelToSnake(s string) string {
	// Insert underscore before uppercase letters
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(snake)
}

// SnakeToCamel converts snake_case to camelCase
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// IsSkippedField checks if a field should be skipped in coverage analysis
func IsSkippedField(fieldName string) bool {
	// Check both the original name and snake_case version
	if SkippedFields[fieldName] {
		return true
	}
	if SkippedFields[CamelToSnake(fieldName)] {
		return true
	}

	// Skip nested array/object fields (complex structures need separate handling)
	// e.g., "sections[].services[]", "subscribe.email", "authentication.password_protection"
	if strings.Contains(fieldName, "[].") || strings.Contains(fieldName, ".") {
		return true
	}

	return false
}

// NormalizeTypeName normalizes type names for comparison
func NormalizeTypeName(typeName string) string {
	typeName = strings.ToLower(strings.TrimSpace(typeName))

	// Handle common variations
	switch typeName {
	case "integer", "int", "int64", "int32", "number":
		return "number"
	case "string", "str":
		return "string"
	case "boolean", "bool":
		return "boolean"
	case "array", "list", "[]string", "[]int":
		return "array"
	case "object", "map", "struct":
		return "object"
	case "enum":
		return "enum"
	}

	// Handle enum<string>, array<string>, etc.
	if strings.HasPrefix(typeName, "enum") {
		return "enum"
	}
	if strings.HasPrefix(typeName, "array") || strings.HasPrefix(typeName, "list") {
		return "array"
	}

	return typeName
}

// ExtractAPIEndpointSection extracts the section from an endpoint filename
// e.g., "monitors_create.json" -> "monitors"
// e.g., "statuspages_subscribers_create.json" -> "statuspages_subscribers"
func ExtractAPIEndpointSection(filename string) string {
	base := strings.TrimSuffix(filename, ".json")

	// Handle nested endpoints like "statuspages_subscribers_create"
	nestedPrefixes := []string{"statuspages_subscribers"}
	for _, prefix := range nestedPrefixes {
		if strings.HasPrefix(base, prefix) {
			return prefix
		}
	}

	// Default: take first part
	parts := strings.Split(base, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// ExtractAPIEndpointOperation extracts the operation from an endpoint filename
// e.g., "monitors_create.json" -> "create"
func ExtractAPIEndpointOperation(filename string) string {
	base := strings.TrimSuffix(filename, ".json")
	parts := strings.Split(base, "_")
	if len(parts) > 1 {
		return strings.Join(parts[1:], "_")
	}
	return ""
}

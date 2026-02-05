// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package contract

import (
	"fmt"
	"regexp"
	"strings"
)

// FieldDiscovery represents a field discovered through contract testing
// that differs from what the API documentation shows.
type FieldDiscovery struct {
	Endpoint   string      `json:"endpoint"`    // e.g., "POST /v1/healthchecks"
	FieldName  string      `json:"field_name"`  // e.g., "lastLogStartDate"
	Source     FieldSource `json:"source"`      // request, response, or both
	Type       string      `json:"type"`        // inferred type from cassette
	DocStatus  DocStatus   `json:"doc_status"`  // undocumented, documented, differs
	DocType    string      `json:"doc_type"`    // type from documentation (if any)
	Suggestion string      `json:"suggestion"`  // action suggestion
}

// DocStatus indicates how a field relates to documentation.
type DocStatus string

const (
	DocStatusUndocumented DocStatus = "undocumented" // Field in API but not in docs
	DocStatusDocumented   DocStatus = "documented"   // Field in both API and docs
	DocStatusDiffers      DocStatus = "differs"      // Field exists but type differs
	DocStatusDeprecated   DocStatus = "deprecated"   // In docs but not in API response
)

// ComparisonResult holds the result of comparing cassettes to documentation.
type ComparisonResult struct {
	Resource    string           `json:"resource"`     // e.g., "healthchecks"
	Endpoint    string           `json:"endpoint"`     // e.g., "POST /v1/healthchecks"
	Discoveries []FieldDiscovery `json:"discoveries"`  // Fields discovered via contract testing
	Summary     ComparisonStats  `json:"summary"`      // Statistics
}

// ComparisonStats holds comparison statistics.
type ComparisonStats struct {
	DocumentedFields   int `json:"documented_fields"`   // Fields in both API and docs
	UndocumentedFields int `json:"undocumented_fields"` // Fields in API but not docs
	DeprecatedFields   int `json:"deprecated_fields"`   // Fields in docs but not API
	TypeMismatches     int `json:"type_mismatches"`     // Fields with different types
}

// DocumentedField represents a field from API documentation.
type DocumentedField struct {
	Name     string
	Type     string
	Required bool
}

// CompareWithDocumentation compares cassette-extracted schema against documented fields.
func CompareWithDocumentation(cassetteSchema *CassetteSchema, docFields map[string][]DocumentedField) []ComparisonResult {
	var results []ComparisonResult

	for endpointKey, endpoint := range cassetteSchema.Endpoints {
		result := ComparisonResult{
			Resource: extractResourceFromPath(endpoint.Path),
			Endpoint: endpointKey,
		}

		// Get documented fields for this endpoint's resource
		// Build map with normalized names for comparison
		docFieldList := docFields[result.Resource]
		docFieldMap := make(map[string]DocumentedField)
		normalizedDocMap := make(map[string]DocumentedField) // snake_case normalized
		for _, f := range docFieldList {
			docFieldMap[f.Name] = f
			normalizedDocMap[normalizeFieldName(f.Name)] = f
		}

		// Compare response fields (most important for schema discovery)
		for fieldName, cassetteField := range endpoint.ResponseFields {
			discovery := FieldDiscovery{
				Endpoint:  endpointKey,
				FieldName: fieldName,
				Source:    cassetteField.Source,
				Type:      cassetteField.Type,
			}

			// Try exact match first, then normalized (snake_case) match
			docField, exists := docFieldMap[fieldName]
			if !exists {
				normalizedName := normalizeFieldName(fieldName)
				docField, exists = normalizedDocMap[normalizedName]
			}

			if exists {
				// Field is documented
				docNormType := normalizeTypeForComparison(docField.Type)
				cassetteNormType := normalizeTypeForComparison(cassetteField.Type)

				if docNormType != cassetteNormType && docNormType != "unknown" && cassetteNormType != "unknown" {
					discovery.DocStatus = DocStatusDiffers
					discovery.DocType = docField.Type
					discovery.Suggestion = fmt.Sprintf("API returns %s but docs say %s", cassetteField.Type, docField.Type)
					result.Summary.TypeMismatches++
				} else {
					discovery.DocStatus = DocStatusDocumented
					result.Summary.DocumentedFields++
					continue // Skip adding documented fields to discoveries
				}
			} else {
				// Field is undocumented
				discovery.DocStatus = DocStatusUndocumented
				discovery.Suggestion = fmt.Sprintf("Add '%s' (%s) to API documentation", fieldName, cassetteField.Type)
				result.Summary.UndocumentedFields++
			}

			result.Discoveries = append(result.Discoveries, discovery)
		}

		// Check for deprecated fields (in docs but not in API response)
		// Build normalized response field map for comparison
		normalizedRespMap := make(map[string]bool)
		for respFieldName := range endpoint.ResponseFields {
			normalizedRespMap[normalizeFieldName(respFieldName)] = true
		}

		for fieldName := range docFieldMap {
			normalizedDocName := normalizeFieldName(fieldName)
			// Check both exact match and normalized match
			_, exactExists := endpoint.ResponseFields[fieldName]
			_, normalizedExists := normalizedRespMap[normalizedDocName]

			if !exactExists && !normalizedExists {
				// Only flag if we saw the endpoint (had a successful response)
				if len(endpoint.ResponseFields) > 0 {
					discovery := FieldDiscovery{
						Endpoint:   endpointKey,
						FieldName:  fieldName,
						DocStatus:  DocStatusDeprecated,
						DocType:    docFieldMap[fieldName].Type,
						Suggestion: fmt.Sprintf("Field '%s' documented but not returned by API", fieldName),
					}
					result.Discoveries = append(result.Discoveries, discovery)
					result.Summary.DeprecatedFields++
				}
			}
		}

		if len(result.Discoveries) > 0 || result.Summary.DocumentedFields > 0 {
			results = append(results, result)
		}
	}

	return results
}

// normalizeFieldName converts camelCase to snake_case for comparison.
// e.g., "periodValue" -> "period_value", "isPaused" -> "is_paused"
func normalizeFieldName(name string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// extractResourceFromPath extracts the resource name from an API path.
// e.g., "/v1/healthchecks/{id}" -> "healthchecks"
func extractResourceFromPath(path string) string {
	// Remove version prefix
	path = regexp.MustCompile(`^/v\d+/`).ReplaceAllString(path, "")

	// Get first path segment
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part != "" && part != "{id}" {
			return part
		}
	}
	return ""
}

// normalizeTypeForComparison normalizes type names for comparison.
func normalizeTypeForComparison(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "int", "integer", "int64", "int32":
		return "integer"
	case "float", "float64", "double", "number":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "string", "str", "text":
		return "string"
	case "array", "list", "[]":
		return "array"
	case "object", "map", "struct":
		return "object"
	default:
		return t
	}
}

// GenerateDiscoveryReport creates a markdown report of discovered undocumented fields.
func GenerateDiscoveryReport(results []ComparisonResult) string {
	var sb strings.Builder

	sb.WriteString("# API Documentation Gaps (Discovered via Contract Testing)\n\n")
	sb.WriteString("These fields are **returned by the API** but **not documented**.\n\n")

	totalUndocumented := 0
	totalMismatches := 0

	for _, result := range results {
		if len(result.Discoveries) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n", result.Resource))
		sb.WriteString(fmt.Sprintf("**Endpoint:** `%s`\n\n", result.Endpoint))

		// Group by status
		var undocumented, differs, deprecated []FieldDiscovery
		for _, d := range result.Discoveries {
			switch d.DocStatus {
			case DocStatusUndocumented:
				undocumented = append(undocumented, d)
			case DocStatusDiffers:
				differs = append(differs, d)
			case DocStatusDeprecated:
				deprecated = append(deprecated, d)
			}
		}

		if len(undocumented) > 0 {
			sb.WriteString("### Undocumented Fields\n\n")
			for _, d := range undocumented {
				sb.WriteString(fmt.Sprintf("- `%s` (%s) - %s\n", d.FieldName, d.Type, d.Suggestion))
			}
			sb.WriteString("\n")
			totalUndocumented += len(undocumented)
		}

		if len(differs) > 0 {
			sb.WriteString("### Type Mismatches\n\n")
			for _, d := range differs {
				sb.WriteString(fmt.Sprintf("- `%s`: API returns %s, docs say %s\n", d.FieldName, d.Type, d.DocType))
			}
			sb.WriteString("\n")
			totalMismatches += len(differs)
		}

		if len(deprecated) > 0 {
			sb.WriteString("### Possibly Deprecated\n\n")
			for _, d := range deprecated {
				sb.WriteString(fmt.Sprintf("- `%s` (%s) - %s\n", d.FieldName, d.DocType, d.Suggestion))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("**Summary:** %d undocumented fields, %d type mismatches\n", totalUndocumented, totalMismatches))

	return sb.String()
}

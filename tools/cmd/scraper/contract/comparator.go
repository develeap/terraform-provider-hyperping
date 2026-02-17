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
	Endpoint   string      `json:"endpoint"`   // e.g., "POST /v1/healthchecks"
	FieldName  string      `json:"field_name"` // e.g., "lastLogStartDate"
	Source     FieldSource `json:"source"`     // request, response, or both
	Type       string      `json:"type"`       // inferred type from cassette
	DocStatus  DocStatus   `json:"doc_status"` // undocumented, documented, differs
	DocType    string      `json:"doc_type"`   // type from documentation (if any)
	Suggestion string      `json:"suggestion"` // action suggestion
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
	Resource    string           `json:"resource"`    // e.g., "healthchecks"
	Endpoint    string           `json:"endpoint"`    // e.g., "POST /v1/healthchecks"
	Discoveries []FieldDiscovery `json:"discoveries"` // Fields discovered via contract testing
	Summary     ComparisonStats  `json:"summary"`     // Statistics
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

// docFieldMaps groups the two lookup maps built from documented fields.
type docFieldMaps struct {
	exact      map[string]DocumentedField // keyed by original name
	normalized map[string]DocumentedField // keyed by snake_case name
}

// buildDocFieldMaps creates exact and normalized lookup maps from a slice of documented fields.
func buildDocFieldMaps(docFieldList []DocumentedField) docFieldMaps {
	exact := make(map[string]DocumentedField, len(docFieldList))
	normalized := make(map[string]DocumentedField, len(docFieldList))
	for _, f := range docFieldList {
		exact[f.Name] = f
		normalized[normalizeFieldName(f.Name)] = f
	}
	return docFieldMaps{exact: exact, normalized: normalized}
}

// CompareWithDocumentation compares cassette-extracted schema against documented fields.
func CompareWithDocumentation(cassetteSchema *CassetteSchema, docFields map[string][]DocumentedField) []ComparisonResult {
	var results []ComparisonResult

	for endpointKey, endpoint := range cassetteSchema.Endpoints {
		result := ComparisonResult{
			Resource: extractResourceFromPath(endpoint.Path),
			Endpoint: endpointKey,
		}

		maps := buildDocFieldMaps(docFields[result.Resource])

		result = compareResponseFields(result, endpoint.ResponseFields, maps)
		result = detectDeprecatedFields(result, endpoint.ResponseFields, maps.exact)

		if len(result.Discoveries) > 0 || result.Summary.DocumentedFields > 0 {
			results = append(results, result)
		}
	}

	return results
}

// compareResponseFields checks each cassette response field against documentation.
func compareResponseFields(result ComparisonResult, responseFields map[string]APIFieldSchema, maps docFieldMaps) ComparisonResult {
	for fieldName, cassetteField := range responseFields {
		discovery := FieldDiscovery{
			Endpoint:  result.Endpoint,
			FieldName: fieldName,
			Source:    cassetteField.Source,
			Type:      cassetteField.Type,
		}

		docField, exists := lookupDocField(fieldName, maps)
		if exists {
			discovery, result = classifyDocumentedField(discovery, result, docField, cassetteField)
			if discovery.DocStatus == DocStatusDocumented {
				continue
			}
		} else {
			discovery.DocStatus = DocStatusUndocumented
			discovery.Suggestion = fmt.Sprintf("Add '%s' (%s) to API documentation", fieldName, cassetteField.Type)
			result.Summary.UndocumentedFields++
		}

		result.Discoveries = append(result.Discoveries, discovery)
	}
	return result
}

// lookupDocField tries exact then normalized lookup; returns the field and whether it was found.
func lookupDocField(fieldName string, maps docFieldMaps) (DocumentedField, bool) {
	if f, ok := maps.exact[fieldName]; ok {
		return f, true
	}
	if f, ok := maps.normalized[normalizeFieldName(fieldName)]; ok {
		return f, true
	}
	return DocumentedField{}, false
}

// classifyDocumentedField compares types and updates discovery/stats for a field that exists in docs.
func classifyDocumentedField(discovery FieldDiscovery, result ComparisonResult, docField DocumentedField, cassetteField APIFieldSchema) (FieldDiscovery, ComparisonResult) {
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
	}
	return discovery, result
}

// detectDeprecatedFields checks for documented fields absent from the API response.
func detectDeprecatedFields(result ComparisonResult, responseFields map[string]APIFieldSchema, docFieldMap map[string]DocumentedField) ComparisonResult {
	if len(responseFields) == 0 {
		return result
	}

	// Build normalized response-field set once.
	normalizedRespMap := make(map[string]bool, len(responseFields))
	for respFieldName := range responseFields {
		normalizedRespMap[normalizeFieldName(respFieldName)] = true
	}

	for fieldName, docField := range docFieldMap {
		_, exactExists := responseFields[fieldName]
		_, normalizedExists := normalizedRespMap[normalizeFieldName(fieldName)]
		if exactExists || normalizedExists {
			continue
		}
		discovery := FieldDiscovery{
			Endpoint:   result.Endpoint,
			FieldName:  fieldName,
			DocStatus:  DocStatusDeprecated,
			DocType:    docField.Type,
			Suggestion: fmt.Sprintf("Field '%s' documented but not returned by API", fieldName),
		}
		result.Discoveries = append(result.Discoveries, discovery)
		result.Summary.DeprecatedFields++
	}
	return result
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

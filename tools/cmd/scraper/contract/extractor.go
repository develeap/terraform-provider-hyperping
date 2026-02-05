// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package contract

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/tools/cmd/scraper/utils"
	"gopkg.in/yaml.v3"
)

// Cassette represents the structure of a go-vcr cassette file.
type Cassette struct {
	Version      int           `yaml:"version"`
	Interactions []Interaction `yaml:"interactions"`
}

// Interaction represents a single HTTP request/response pair.
type Interaction struct {
	Request  RequestRecord  `yaml:"request"`
	Response ResponseRecord `yaml:"response"`
}

// RequestRecord represents the request portion of an interaction.
type RequestRecord struct {
	Method string            `yaml:"method"`
	URL    string            `yaml:"url"`
	Body   string            `yaml:"body"`
	Form   map[string]string `yaml:"form"`
}

// ResponseRecord represents the response portion of an interaction.
type ResponseRecord struct {
	StatusCode int    `yaml:"code"`
	Body       string `yaml:"body"`
}

// ExtractFromCassettes reads all cassettes in a directory and extracts field schemas.
func ExtractFromCassettes(cassetteDir string) (*CassetteSchema, error) {
	schema := NewCassetteSchema(cassetteDir)

	files, err := filepath.Glob(filepath.Join(cassetteDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob cassette files: %w", err)
	}

	for _, file := range files {
		if err := extractFromCassette(cassetteDir, file, schema); err != nil {
			return nil, fmt.Errorf("failed to extract from %s: %w", file, err)
		}
	}

	return schema, nil
}

// extractFromCassette extracts schemas from a single cassette file.
func extractFromCassette(baseDir, path string, schema *CassetteSchema) error {
	data, err := utils.SafeReadFile(baseDir, path)
	if err != nil {
		return fmt.Errorf("failed to read cassette: %w", err)
	}

	var cassette Cassette
	if err := yaml.Unmarshal(data, &cassette); err != nil {
		return fmt.Errorf("failed to parse cassette: %w", err)
	}

	for _, interaction := range cassette.Interactions {
		if err := extractInteraction(&interaction, schema); err != nil {
			// Log but continue processing
			continue
		}
	}

	return nil
}

// extractInteraction extracts field information from a single interaction.
func extractInteraction(interaction *Interaction, schema *CassetteSchema) error {
	// Normalize path (remove UUIDs, IDs)
	normalizedPath := normalizePath(interaction.Request.URL)
	endpoint := schema.GetOrCreateEndpoint(interaction.Request.Method, normalizedPath)

	// Track status code
	if !contains(endpoint.StatusCodes, interaction.Response.StatusCode) {
		endpoint.StatusCodes = append(endpoint.StatusCodes, interaction.Response.StatusCode)
	}

	// Only process successful responses (2xx) for schema extraction
	if interaction.Response.StatusCode < 200 || interaction.Response.StatusCode >= 300 {
		return nil
	}

	// Extract request fields
	if interaction.Request.Body != "" {
		var reqBody map[string]any
		if err := json.Unmarshal([]byte(interaction.Request.Body), &reqBody); err == nil {
			extractFields(reqBody, endpoint.RequestFields, SourceRequest, "")
		}
	}

	// Extract response fields
	if interaction.Response.Body != "" {
		var respBody any
		if err := json.Unmarshal([]byte(interaction.Response.Body), &respBody); err == nil {
			switch v := respBody.(type) {
			case map[string]any:
				extractFields(v, endpoint.ResponseFields, SourceResponse, "")
			case []any:
				// Array response - extract from first element
				if len(v) > 0 {
					if obj, ok := v[0].(map[string]any); ok {
						extractFields(obj, endpoint.ResponseFields, SourceResponse, "")
					}
				}
			}
		}
	}

	return nil
}

// normalizePath normalizes a URL path by replacing UUIDs and IDs with placeholders.
func normalizePath(rawURL string) string {
	// Remove query string
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		rawURL = rawURL[:idx]
	}

	// Remove host if present
	path := rawURL
	if strings.HasPrefix(rawURL, "http") {
		// Find path after host (skip http(s)://host)
		if idx := strings.Index(rawURL[8:], "/"); idx != -1 {
			path = rawURL[8+idx:]
		}
	}

	// Split path into segments
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		// Replace Hyperping-style IDs: prefix_alphanumeric (e.g., hc_abc123, mon_xyz789)
		if matched, _ := regexp.MatchString(`^[a-z]+_[A-Za-z0-9]+$`, seg); matched {
			segments[i] = "{id}"
		}
	}

	return strings.Join(segments, "/")
}

// extractFields recursively extracts field schemas from a JSON object.
func extractFields(obj map[string]any, fields map[string]APIFieldSchema, source FieldSource, prefix string) {
	for name, value := range obj {
		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}

		existing, exists := fields[fullName]
		if !exists {
			existing = APIFieldSchema{
				Name:     fullName,
				Source:   source,
				Examples: []any{},
			}
		} else {
			// Merge source
			existing.Source |= source
		}

		// Infer type
		inferredType, childType, isNull := inferType(value)
		if isNull {
			existing.Nullable = true
		} else {
			existing.Type = inferredType
			existing.ChildType = childType
		}

		// Add example (up to 3)
		if len(existing.Examples) < 3 && value != nil {
			// Don't store large objects/arrays as examples
			if inferredType != "object" && inferredType != "array" {
				existing.Examples = append(existing.Examples, value)
			}
		}

		fields[fullName] = existing

		// Recurse into nested objects
		if nested, ok := value.(map[string]any); ok {
			extractFields(nested, fields, source, fullName)
		}

		// Handle arrays of objects
		if arr, ok := value.([]any); ok && len(arr) > 0 {
			if nestedObj, ok := arr[0].(map[string]any); ok {
				extractFields(nestedObj, fields, source, fullName+"[]")
			}
		}
	}
}

// inferType infers the JSON schema type from a value.
func inferType(value any) (typeName string, childType string, isNull bool) {
	switch v := value.(type) {
	case nil:
		return "", "", true
	case string:
		return "string", "", false
	case float64:
		if v == float64(int(v)) {
			return "integer", "", false
		}
		return "number", "", false
	case bool:
		return "boolean", "", false
	case []any:
		if len(v) > 0 {
			childTypeName, _, _ := inferType(v[0])
			return "array", childTypeName, false
		}
		return "array", "", false
	case map[string]any:
		return "object", "", false
	default:
		return "unknown", "", false
	}
}

// contains checks if a slice contains a value.
func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

package extractor

import (
	"regexp"
	"strings"
)

// isValidParameterName checks if a string looks like a parameter name
func isValidParameterName(name string) bool {
	if len(name) == 0 || len(name) > 100 {
		return false
	}

	// Must start with letter or underscore
	if !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(name) {
		return false
	}

	// Should only contain alphanumeric, underscore, or dash
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return false
	}

	// Filter out common false positives
	lower := strings.ToLower(name)
	falsePositives := []string{
		"parameter", "type", "required", "optional", "description",
		"name", "value", "example", "default", "note", "endpoint",
	}
	for _, fp := range falsePositives {
		if lower == fp {
			return false
		}
	}

	return true
}

// isLikelyParameter filters out common non-parameter JSON fields
func isLikelyParameter(name string) bool {
	// Filter out metadata fields
	nonParams := []string{
		"error", "message", "status", "code", "data", "meta",
		"timestamp", "created_at", "updated_at", "id", "uuid",
	}

	lower := strings.ToLower(name)
	for _, np := range nonParams {
		if lower == np {
			return false
		}
	}

	return isValidParameterName(name)
}

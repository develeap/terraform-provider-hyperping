package main

import (
	"log"
	"regexp"
	"strings"
)

// ExtractAPIParameters extracts structured parameter information from scraped HTML
// This is a best-effort extraction based on common documentation patterns
func ExtractAPIParameters(pageData *PageData) []APIParameter {
	var params []APIParameter

	// Strategy: Look for parameter definitions in the text content
	// Hyperping docs typically format parameters as:
	// - name (type, required/optional) - description
	// - name: description
	// - Parameter: name, Type: string, Required: true

	text := pageData.Text

	// Pattern 1: Look for parameter tables or lists
	// Common patterns:
	// "name string required Description here"
	// "name (string, required) - Description"
	// "• name - description"

	params = append(params, extractFromBulletPoints(text)...)
	params = append(params, extractFromTables(text)...)
	params = append(params, extractFromCodeBlocks(pageData.HTML)...)

	// Deduplicate parameters by name
	seen := make(map[string]bool)
	unique := []APIParameter{}
	for _, p := range params {
		if !seen[p.Name] {
			unique = append(unique, p)
			seen[p.Name] = true
		}
	}

	return unique
}

// extractFromBulletPoints looks for Hyperping's specific parameter format
// Format: "namestringrequired\nDescription text\nExample:..."
func extractFromBulletPoints(text string) []APIParameter {
	var params []APIParameter

	lines := strings.Split(text, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Look for Hyperping's parameter format:
		// "namestringrequired" or "nameintegeroptional" etc.
		// Pattern: lowercase_name + type + required/optional (all concatenated)
		param := parseHyperpingParameter(line)
		if param.Name != "" {
			// Try to get description from next line
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if !strings.HasPrefix(nextLine, "Example:") &&
					!strings.HasPrefix(nextLine, "Available options:") &&
					!strings.HasPrefix(nextLine, "Default:") &&
					len(nextLine) > 0 {
					param.Description = nextLine
				}
			}
			params = append(params, param)
		}
	}

	return params
}

// parseHyperpingParameter parses Hyperping's concatenated parameter format
// Example: "namestringrequired" → name="name", type="string", required=true
// Example: "check_frequencynumberoptional" → name="check_frequency", type="number", required=false
func parseHyperpingParameter(line string) APIParameter {
	param := APIParameter{}

	lower := strings.ToLower(line)

	// Pattern: parameter_name + type + required/optional
	// Types: string, number, boolean, array, object, enum

	// Try to find type keywords
	types := []string{"string", "number", "integer", "boolean", "array", "object", "enum"}
	var typeFound string
	var typePos int = -1

	for _, t := range types {
		pos := strings.Index(lower, t)
		if pos > 0 { // Must not be at start (that would be weird)
			if typePos == -1 || pos < typePos {
				typePos = pos
				typeFound = t
			}
		}
	}

	if typePos == -1 {
		return param // No type found
	}

	// Extract name (everything before type)
	name := line[:typePos]
	if !isValidParameterName(name) {
		return param
	}
	param.Name = name
	param.Type = typeFound

	// Check what comes after the type
	afterType := lower[typePos+len(typeFound):]

	// Look for required/optional
	if strings.Contains(afterType, "required") {
		param.Required = true
	} else if strings.Contains(afterType, "optional") {
		param.Required = false
	}

	// Handle enum types specially
	if strings.HasPrefix(afterType, "<string>") || strings.HasPrefix(afterType, "<number>") {
		param.Type = "enum"
	}

	return param
}

// extractFromTables looks for parameter tables in text
// Tables often have headers like: Parameter | Type | Required | Description
func extractFromTables(text string) []APIParameter {
	var params []APIParameter

	lines := strings.Split(text, "\n")

	// State machine to detect tables
	inTable := false
	headerFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect table headers
		if strings.Contains(strings.ToLower(line), "parameter") &&
			(strings.Contains(strings.ToLower(line), "type") ||
				strings.Contains(strings.ToLower(line), "required")) {
			inTable = true
			headerFound = true
			continue
		}

		// Skip separator lines
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "===") {
			continue
		}

		// Process table rows
		if inTable && headerFound && len(line) > 0 {
			// Look for pipe-separated values or tab-separated
			if strings.Contains(line, "|") || strings.Contains(line, "\t") {
				param := parseTableRow(line)
				if param.Name != "" {
					params = append(params, param)
				}
			} else if len(line) > 0 && !strings.Contains(strings.ToLower(line), "parameter") {
				// Table might have ended
				inTable = false
				headerFound = false
			}
		}
	}

	return params
}

// extractFromCodeBlocks extracts parameters from JSON examples in HTML
func extractFromCodeBlocks(html string) []APIParameter {
	var params []APIParameter

	// Look for JSON request/response examples
	// Extract field names from JSON structure
	re := regexp.MustCompile(`"(\w+)"\s*:`)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			fieldName := match[1]

			// Filter out common non-parameter fields
			if isLikelyParameter(fieldName) {
				params = append(params, APIParameter{
					Name: fieldName,
					Type: "unknown", // Can't determine type from field name alone
				})
			}
		}
	}

	return params
}

// parseParameterLine attempts to parse a single line as a parameter definition
// Handles formats like:
// - "name (string, required) - Description"
// - "name string - Description"
// - "name - Description"
func parseParameterLine(line string) APIParameter {
	param := APIParameter{}

	// Try to extract name from start of line
	// Name is usually the first word before parentheses or dash
	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 0 {
		return param
	}

	// Extract name (first word, cleaned)
	name := strings.Trim(parts[0], "`:.,")
	if !isValidParameterName(name) {
		return param
	}
	param.Name = name

	// Look for type and required information in parentheses
	if strings.Contains(line, "(") && strings.Contains(line, ")") {
		re := regexp.MustCompile(`\((.*?)\)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			info := strings.ToLower(matches[1])

			// Extract type
			if strings.Contains(info, "string") {
				param.Type = "string"
			} else if strings.Contains(info, "boolean") || strings.Contains(info, "bool") {
				param.Type = "boolean"
			} else if strings.Contains(info, "integer") || strings.Contains(info, "int") {
				param.Type = "integer"
			} else if strings.Contains(info, "array") {
				param.Type = "array"
			} else if strings.Contains(info, "object") {
				param.Type = "object"
			}

			// Check if required
			param.Required = strings.Contains(info, "required")
		}
	}

	// Look for type without parentheses (e.g., "name string required")
	if param.Type == "" && len(parts) > 1 {
		rest := strings.ToLower(parts[1])
		if strings.Contains(rest, "string") {
			param.Type = "string"
		} else if strings.Contains(rest, "boolean") || strings.Contains(rest, "bool") {
			param.Type = "boolean"
		} else if strings.Contains(rest, "integer") || strings.Contains(rest, "int") {
			param.Type = "integer"
		} else if strings.Contains(rest, "array") {
			param.Type = "array"
		} else if strings.Contains(rest, "object") {
			param.Type = "object"
		}

		if strings.Contains(rest, "required") {
			param.Required = true
		}
	}

	// Extract description (text after dash or parentheses)
	descStart := strings.Index(line, " - ")
	if descStart > 0 {
		param.Description = strings.TrimSpace(line[descStart+3:])
	}

	return param
}

// parseTableRow parses a table row into a parameter
// Expected format: name | type | required | description
func parseTableRow(line string) APIParameter {
	param := APIParameter{}

	// Split by pipe or tab
	var parts []string
	if strings.Contains(line, "|") {
		parts = strings.Split(line, "|")
	} else {
		parts = strings.Split(line, "\t")
	}

	if len(parts) < 2 {
		return param
	}

	// Clean up parts
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Extract name (first column)
	name := strings.Trim(parts[0], "`:.,")
	if !isValidParameterName(name) {
		return param
	}
	param.Name = name

	// Extract type (second column)
	if len(parts) > 1 {
		param.Type = inferType(parts[1])
	}

	// Extract required (third column)
	if len(parts) > 2 {
		param.Required = strings.Contains(strings.ToLower(parts[2]), "required") ||
			strings.Contains(strings.ToLower(parts[2]), "yes") ||
			strings.Contains(strings.ToLower(parts[2]), "true")
	}

	// Extract description (fourth column)
	if len(parts) > 3 {
		param.Description = parts[3]
	}

	return param
}

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

// inferType attempts to infer parameter type from a string
func inferType(typeStr string) string {
	lower := strings.ToLower(typeStr)

	if strings.Contains(lower, "string") || strings.Contains(lower, "text") {
		return "string"
	}
	if strings.Contains(lower, "bool") {
		return "boolean"
	}
	if strings.Contains(lower, "int") || strings.Contains(lower, "number") {
		return "integer"
	}
	if strings.Contains(lower, "array") || strings.Contains(lower, "list") {
		return "array"
	}
	if strings.Contains(lower, "object") || strings.Contains(lower, "map") {
		return "object"
	}

	return typeStr
}

// LogParameterExtractionResults logs the extracted parameters for debugging
func LogParameterExtractionResults(url string, params []APIParameter) {
	if len(params) == 0 {
		log.Printf("      📋 No parameters extracted\n")
		return
	}

	log.Printf("      📋 Extracted %d parameters:\n", len(params))
	for i, p := range params {
		required := "optional"
		if p.Required {
			required = "required"
		}
		log.Printf("         %d. %s (%s, %s)\n", i+1, p.Name, p.Type, required)
	}
}

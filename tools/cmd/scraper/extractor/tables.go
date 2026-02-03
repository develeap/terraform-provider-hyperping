package extractor

import "strings"

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

package extractor

import (
	"strings"
)

// extractFromHyperpingFormat parses Hyperping's specific parameter format
// Format: "namestringrequired\nDescription text\nExample:..."
func extractFromHyperpingFormat(text string) []APIParameter {
	var params []APIParameter

	lines := strings.Split(text, "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Parse Hyperping's concatenated parameter format
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
// Examples:
//   - "namestringrequired" → name="name", type="string", required=true
//   - "check_frequencynumberoptional" → name="check_frequency", type="number", required=false
func parseHyperpingParameter(line string) APIParameter {
	param := APIParameter{}

	lower := strings.ToLower(line)

	// Known type keywords in priority order
	types := []string{"string", "number", "integer", "boolean", "array", "object", "enum"}

	// Find first type keyword
	var typeFound string
	var typePos int = -1

	for _, t := range types {
		pos := strings.Index(lower, t)
		if pos > 0 { // Must not be at start
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

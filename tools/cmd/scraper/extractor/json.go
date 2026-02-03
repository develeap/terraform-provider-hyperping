package extractor

import "regexp"

// extractFromJSONExamples extracts parameters from JSON examples in HTML
func extractFromJSONExamples(html string) []APIParameter {
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

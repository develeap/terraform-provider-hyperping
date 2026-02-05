package extractor

import (
	"regexp"
	"strings"
)

// RequestParameterSections defines which section titles contain request parameters
// These are the ONLY sections we should extract parameters from
var RequestParameterSections = map[string]bool{
	"body":             true,
	"path parameters":  true,
	"query parameters": true,
	"headers":          true,
	"request body":     true,
}

// ResponseSectionPatterns defines patterns that indicate response/object documentation
// We should NOT extract parameters from these sections
var ResponseSectionPatterns = []string{
	"object",
	"fields",
	"response",
	"reference",
	"options",
}

// isRequestParameterSection checks if a section title indicates request parameters
func isRequestParameterSection(title string) bool {
	lower := strings.ToLower(strings.TrimSpace(title))

	// Check if it's explicitly a request section
	if RequestParameterSections[lower] {
		return true
	}

	// Check if it contains response/object patterns (exclude these)
	for _, pattern := range ResponseSectionPatterns {
		if strings.Contains(lower, pattern) {
			return false
		}
	}

	// Default to false for unknown sections
	return false
}

// ExtractFromHTML extracts API parameters from the structured HTML
// This is the primary extraction method - it uses Hyperping's semantic HTML structure
func ExtractFromHTML(html string) []APIParameter {
	var params []APIParameter

	// Strategy: Find all api-param-row elements and determine their section by looking backwards
	// This is more robust than trying to match complex nested div patterns

	// First, find all section titles and their positions
	titlePattern := regexp.MustCompile(`<span[^>]*class="[^"]*api-params-title[^"]*"[^>]*>([^<]+)</span>`)
	titleMatches := titlePattern.FindAllStringSubmatchIndex(html, -1)

	// Build a list of section boundaries
	type section struct {
		title     string
		startPos  int
		endPos    int
		isRequest bool
	}
	var sections []section

	for i, match := range titleMatches {
		if len(match) < 4 {
			continue
		}
		title := html[match[2]:match[3]]
		startPos := match[0]

		// End position is the start of next section, or end of HTML
		endPos := len(html)
		if i+1 < len(titleMatches) {
			endPos = titleMatches[i+1][0]
		}

		sections = append(sections, section{
			title:     title,
			startPos:  startPos,
			endPos:    endPos,
			isRequest: isRequestParameterSection(title),
		})
	}

	// Now extract parameters only from request sections
	rowPattern := regexp.MustCompile(`(?s)<div[^>]*class="[^"]*api-param-row[^"]*"[^>]*>(.*?)</div>\s*</div>`)

	for _, sec := range sections {
		if !sec.isRequest {
			continue
		}

		// Extract content for this section
		sectionHTML := html[sec.startPos:sec.endPos]

		// Find all param rows in this section
		rows := rowPattern.FindAllStringSubmatch(sectionHTML, -1)
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			param := extractParamFromRow(row[1])
			if param.Name != "" {
				params = append(params, param)
			}
		}
	}

	return params
}

// extractParamsFromTable extracts parameters from a single api-params-table
func extractParamsFromTable(tableHTML string) []APIParameter {
	var params []APIParameter

	// Find all api-param-row elements
	// Each row contains: name, type, required/optional, description, default, options
	rowPattern := regexp.MustCompile(`(?s)<div[^>]*class="[^"]*api-param-row[^"]*"[^>]*>(.*?)</div>\s*</div>`)
	rows := rowPattern.FindAllStringSubmatch(tableHTML, -1)

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		rowContent := row[1]

		param := extractParamFromRow(rowContent)
		if param.Name != "" {
			params = append(params, param)
		}
	}

	return params
}

// extractParamFromRow extracts a single parameter from an api-param-row
func extractParamFromRow(rowHTML string) APIParameter {
	param := APIParameter{}

	// Extract name
	namePattern := regexp.MustCompile(`<span[^>]*class="[^"]*api-param-name[^"]*"[^>]*>([^<]+)</span>`)
	if match := namePattern.FindStringSubmatch(rowHTML); len(match) > 1 {
		param.Name = strings.TrimSpace(match[1])
	}

	if param.Name == "" {
		return param
	}

	// Extract type
	typePattern := regexp.MustCompile(`<span[^>]*class="[^"]*api-param-type[^"]*"[^>]*>([^<]+)</span>`)
	if match := typePattern.FindStringSubmatch(rowHTML); len(match) > 1 {
		param.Type = normalizeType(strings.TrimSpace(match[1]))
	}

	// Check if required
	if strings.Contains(rowHTML, `api-param-required`) {
		param.Required = true
	}

	// Extract description
	descPattern := regexp.MustCompile(`<div[^>]*class="[^"]*api-param-description[^"]*"[^>]*>([^<]+)`)
	if match := descPattern.FindStringSubmatch(rowHTML); len(match) > 1 {
		param.Description = strings.TrimSpace(match[1])
	}

	// Extract default value
	defaultPattern := regexp.MustCompile(`<div[^>]*class="[^"]*api-param-default[^"]*"[^>]*>Default:\s*<code>([^<]+)</code>`)
	if match := defaultPattern.FindStringSubmatch(rowHTML); len(match) > 1 {
		param.Default = strings.TrimSpace(match[1])
	}

	// Extract valid values/options
	optionsPattern := regexp.MustCompile(`<span[^>]*class="[^"]*api-param-option[^"]*"[^>]*>([^<]+)</span>`)
	optionMatches := optionsPattern.FindAllStringSubmatch(rowHTML, -1)
	for _, match := range optionMatches {
		if len(match) > 1 {
			param.ValidValues = append(param.ValidValues, strings.TrimSpace(match[1]))
		}
	}

	return param
}

// HasParameterTables checks if the HTML contains any api-params-table structures
// This is used to determine if we should use text fallback extraction
func HasParameterTables(html string) bool {
	return strings.Contains(html, `class="api-params-table"`) ||
		strings.Contains(html, `class="api-param-row"`)
}

// normalizeType converts HTML type strings to standard types
func normalizeType(typeStr string) string {
	lower := strings.ToLower(typeStr)

	// Handle enum types like "enum<string>"
	if strings.HasPrefix(lower, "enum") {
		return "enum"
	}

	// Handle array types like "array<string>"
	if strings.HasPrefix(lower, "array") {
		return "array"
	}

	// Handle object types
	if strings.Contains(lower, "object") {
		return "object"
	}

	// Standard types
	switch {
	case strings.Contains(lower, "string"):
		return "string"
	case strings.Contains(lower, "number"), strings.Contains(lower, "integer"), strings.Contains(lower, "int"):
		return "number"
	case strings.Contains(lower, "bool"):
		return "boolean"
	default:
		return typeStr
	}
}

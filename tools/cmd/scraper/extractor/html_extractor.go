package extractor

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// RequestParameterSections defines which section titles contain request parameters.
// These are the ONLY sections we should extract parameters from.
var RequestParameterSections = map[string]bool{
	"body":             true,
	"path parameters":  true,
	"query parameters": true,
	"headers":          true,
	"request body":     true,
}

// ResponseSectionPatterns defines patterns that indicate response/object documentation.
// We should NOT extract parameters from these sections.
var ResponseSectionPatterns = []string{
	"object",
	"fields",
	"response",
	"reference",
	"options",
}

// isRequestParameterSection returns true when a section title describes request params.
func isRequestParameterSection(title string) bool {
	lower := strings.ToLower(strings.TrimSpace(title))
	if RequestParameterSections[lower] {
		return true
	}
	for _, pattern := range ResponseSectionPatterns {
		if strings.Contains(lower, pattern) {
			return false
		}
	}
	return false
}

// ExtractFromHTML extracts API parameters from page HTML using GoQuery CSS selectors.
// It only extracts parameters found inside request-parameter sections.
func ExtractFromHTML(html string) []APIParameter {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}
	return ExtractParams(doc)
}

// ExtractParams extracts parameters from an already-parsed goquery.Document.
// It processes each .api-params-table independently, skipping those whose
// title indicates a response/reference section.
func ExtractParams(doc *goquery.Document) []APIParameter {
	var params []APIParameter
	hasTitle := false

	doc.Find(".api-params-table").Each(func(_ int, table *goquery.Selection) {
		title := strings.TrimSpace(table.Find(".api-params-title").First().Text())
		if title != "" {
			hasTitle = true
			if !isRequestParameterSection(title) {
				return // skip response/object sections
			}
		}

		table.Find(".api-param-row").Each(func(_ int, row *goquery.Selection) {
			p := extractParamFromSelection(row)
			if p.Name != "" {
				params = append(params, p)
			}
		})
	})

	// Fallback: page with no section titles but raw rows present.
	if len(params) == 0 && !hasTitle {
		doc.Find(".api-param-row").Each(func(_ int, row *goquery.Selection) {
			p := extractParamFromSelection(row)
			if p.Name != "" {
				params = append(params, p)
			}
		})
	}

	return params
}

// extractParamFromSelection extracts a single APIParameter from an .api-param-row element.
func extractParamFromSelection(row *goquery.Selection) APIParameter {
	p := APIParameter{}

	p.Name = strings.TrimSpace(row.Find(".api-param-name").First().Text())
	if p.Name == "" {
		return p
	}

	p.Type = normalizeType(strings.TrimSpace(row.Find(".api-param-type").First().Text()))
	p.Required = row.Find(".api-param-required").Length() > 0
	p.Description = strings.TrimSpace(row.Find(".api-param-description").First().Text())
	p.Deprecated = row.HasClass("api-param-deprecated") || row.Find(".api-param-deprecated").Length() > 0

	// Default value: look for a code element inside the default container.
	p.Default = strings.TrimSpace(row.Find(".api-param-default code").First().Text())
	if p.Default == "" {
		p.Default = nil
	}

	// Valid enum values.
	row.Find(".api-param-option").Each(func(_ int, s *goquery.Selection) {
		if v := strings.TrimSpace(s.Text()); v != "" {
			p.ValidValues = append(p.ValidValues, v)
		}
	})

	return p
}

// HasParameterTables returns true if the document contains API parameter markup.
func HasParameterTables(doc *goquery.Document) bool {
	return doc.Find(".api-params-table, .api-param-row").Length() > 0
}

// HasParameterTablesInHTML is a convenience wrapper that accepts a raw HTML string.
func HasParameterTablesInHTML(html string) bool {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return strings.Contains(html, "api-param-row")
	}
	return HasParameterTables(doc)
}

// normalizeType converts Hyperping HTML type strings to standard OAS type names.
func normalizeType(typeStr string) string {
	lower := strings.ToLower(typeStr)

	if strings.HasPrefix(lower, "enum") {
		return "enum"
	}
	if strings.HasPrefix(lower, "array") {
		return "array"
	}
	if strings.Contains(lower, "object") {
		return "object"
	}

	switch {
	case strings.Contains(lower, "string"):
		return "string"
	case strings.Contains(lower, "number"),
		strings.Contains(lower, "integer"),
		strings.Contains(lower, "int"):
		return "number"
	case strings.Contains(lower, "bool"):
		return "boolean"
	default:
		return typeStr
	}
}

// extractFromHyperpingFormat is a text-based fallback for pages without structured tables.
// Kept for compatibility with the text fallback path in extractor.go.
func extractFromHyperpingFormat(text string) []APIParameter {
	// Minimal text fallback: find "name (type)" patterns.
	var params []APIParameter
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip obvious non-parameter lines.
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		// Very conservative: only pick up lines with clear param indicators.
		if !strings.Contains(line, "required") && !strings.Contains(line, "optional") {
			continue
		}
		p := APIParameter{
			Name:        strings.Fields(line)[0],
			Description: line,
		}
		if p.Name != "" {
			params = append(params, p)
		}
	}
	return params
}

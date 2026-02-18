// Package extractor provides API parameter extraction from documentation
package extractor

import "log"

// PageData represents scraped content from a single API documentation page
type PageData struct {
	URL       string `json:"url"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	HTML      string `json:"html"`
	Timestamp string `json:"timestamp"`
}

// APIParameter represents a single API parameter extracted from documentation
type APIParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, boolean, integer, array, object
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
	ValidValues []string    `json:"valid_values,omitempty"`
	Deprecated  bool        `json:"deprecated,omitempty"`
}

// ExtractAPIParameters extracts structured parameter information from scraped HTML
// Uses HTML structure parsing for accurate extraction of REQUEST parameters only
func ExtractAPIParameters(pageData *PageData) []APIParameter {
	// Primary strategy: Parse HTML structure directly
	// This is the most reliable method as it uses Hyperping's semantic HTML classes
	// and only extracts from request parameter sections (Body, Path Parameters, Query Parameters)
	// It explicitly ignores response/object documentation sections
	params := ExtractFromHTML(pageData.HTML)

	// If HTML extraction found parameters, use those exclusively
	if len(params) > 0 {
		return deduplicateParameters(params)
	}

	// If the page has parameter tables but none are request sections,
	// don't fall back to text extraction (which would extract response fields)
	if HasParameterTablesInHTML(pageData.HTML) {
		return []APIParameter{} // Return empty - no request parameters on this page
	}

	// Fallback: Try text-based extraction ONLY for pages without standard HTML structure
	// This handles edge cases like authentication pages without structured tables
	params = extractFromHyperpingFormat(pageData.Text)

	return deduplicateParameters(params)
}

// deduplicateParameters removes duplicate parameters by name (keeps first occurrence)
func deduplicateParameters(params []APIParameter) []APIParameter {
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

// LogExtractionResults logs the extracted parameters for debugging
func LogExtractionResults(url string, params []APIParameter) {
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

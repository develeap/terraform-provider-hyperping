// Package extractor provides API parameter extraction from documentation
package extractor

import (
	"log"
)

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
// Uses multiple extraction strategies for robustness
func ExtractAPIParameters(pageData *PageData) []APIParameter {
	var params []APIParameter

	text := pageData.Text

	// Strategy 1: Hyperping's specific format (primary)
	params = append(params, extractFromHyperpingFormat(text)...)

	// Strategy 2: Table formats (backup)
	params = append(params, extractFromTables(text)...)

	// Strategy 3: JSON code blocks (supplementary)
	params = append(params, extractFromJSONExamples(pageData.HTML)...)

	// Deduplicate by name
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

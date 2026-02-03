package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// TestParameterExtraction tests the extraction on a single scraped page
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_extraction.go <json_file>")
		fmt.Println("\nExample: go run test_extraction.go docs_scraped/monitors_create.json")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load page data
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("❌ Failed to read file: %v\n", err)
		os.Exit(1)
	}

	var pageData PageData
	if err := json.Unmarshal(data, &pageData); err != nil {
		fmt.Printf("❌ Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📄 File: %s\n", filename)
	fmt.Printf("🔗 URL: %s\n", pageData.URL)
	fmt.Printf("📏 Content size: %d chars\n\n", len(pageData.Text))

	// Extract parameters
	params := ExtractAPIParameters(&pageData)

	fmt.Printf("📋 Extracted %d parameters:\n\n", len(params))

	if len(params) == 0 {
		fmt.Println("⚠️  No parameters extracted. This could mean:")
		fmt.Println("   - The page has no parameters")
		fmt.Println("   - The extraction patterns need adjustment")
		fmt.Println("   - The page format is different than expected")
		return
	}

	// Print parameters in a table format
	fmt.Printf("%-25s %-12s %-10s %s\n", "NAME", "TYPE", "REQUIRED", "DESCRIPTION")
	fmt.Println("────────────────────────────────────────────────────────────────────────────")

	for _, p := range params {
		required := "optional"
		if p.Required {
			required = "required"
		}

		desc := p.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		fmt.Printf("%-25s %-12s %-10s %s\n", p.Name, p.Type, required, desc)
	}

	fmt.Println("\n✅ Extraction test complete")
}

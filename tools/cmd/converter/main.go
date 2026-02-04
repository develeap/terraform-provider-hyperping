package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PageData struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Text  string `json:"text"`
	HTML  string `json:"html"`
}

func main() {
	inputDir := "tools/cmd/scraper/docs_scraped"
	outputDir := filepath.Join(inputDir, "markdown")

	if err := os.MkdirAll(outputDir, 0750); err != nil {
		log.Fatal(err)
	}

	// Find all JSON files
	files, err := filepath.Glob(filepath.Join(inputDir, "*.json"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("📄 Converting JSON to Markdown...")
	for _, jsonFile := range files {
		baseName := strings.TrimSuffix(filepath.Base(jsonFile), ".json")
		mdFile := filepath.Join(outputDir, baseName+".md")

		if err := convertToMarkdown(jsonFile, mdFile); err != nil {
			fmt.Printf("  ❌ Error converting %s: %v\n", jsonFile, err)
			continue
		}

		fmt.Printf("  ✅ %s.md\n", baseName)
	}

	fmt.Printf("\n📂 Markdown files saved to: %s/\n", outputDir)
}

func convertToMarkdown(jsonFile, mdFile string) error {
	// Read JSON
	data, err := os.ReadFile(jsonFile) // #nosec G304 -- jsonFile is from internal glob
	if err != nil {
		return err
	}

	var page PageData
	if err := json.Unmarshal(data, &page); err != nil {
		return err
	}

	// Create markdown
	var md strings.Builder
	md.WriteString(fmt.Sprintf("# %s\n\n", page.Title))
	md.WriteString(fmt.Sprintf("**Source:** %s\n\n", page.URL))
	md.WriteString("---\n\n")
	md.WriteString(page.Text)
	md.WriteString("\n")

	// Write markdown file
	return os.WriteFile(mdFile, []byte(md.String()), 0600)
}

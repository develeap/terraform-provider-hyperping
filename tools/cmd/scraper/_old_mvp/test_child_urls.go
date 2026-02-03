package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// Test URLs from user's examples
var testURLs = []string{
	"https://hyperping.com/docs/api/statuspages/list",
	"https://hyperping.com/docs/api/statuspages/get",
	"https://hyperping.com/docs/api/statuspages/create",
	"https://hyperping.com/docs/api/statuspages/update",
	"https://hyperping.com/docs/api/statuspages/delete",
	"https://hyperping.com/docs/api/statuspages/subscribers/list",
	"https://hyperping.com/docs/api/statuspages/subscribers/create",
	"https://hyperping.com/docs/api/statuspages/subscribers/delete",
	"https://hyperping.com/docs/api/monitors/list",
	"https://hyperping.com/docs/api/monitors/get",
	"https://hyperping.com/docs/api/monitors/create",
}

func main() {
	fmt.Println("🔍 Testing child URL accessibility...")

	ctx := context.Background()

	// Launch browser
	launchURL := launcher.New().MustLaunch()
	browser := rod.New().ControlURL(launchURL).MustConnect()
	defer browser.MustClose()

	for _, url := range testURLs {
		fmt.Printf("\n📄 Testing: %s\n", url)

		page := browser.MustPage(url)
		defer page.MustClose()

		// Wait for page load
		if err := page.WaitLoad(); err != nil {
			fmt.Printf("  ❌ Failed to load: %v\n", err)
			continue
		}

		// Wait for content to render
		time.Sleep(2 * time.Second)

		// Check for common content indicators
		title, _ := page.MustElement("title").Text()
		fmt.Printf("  Title: %s\n", title)

		// Check for HTTP method indicators
		hasMethod := false
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
		for _, method := range methods {
			if page.MustHas(fmt.Sprintf("text=%s", method)) {
				fmt.Printf("  ✅ Contains HTTP method: %s\n", method)
				hasMethod = true
				break
			}
		}

		// Check for parameter section
		if page.MustHas("text=Parameters") || page.MustHas("text=Request Body") || page.MustHas("text=Query Parameters") {
			fmt.Printf("  ✅ Contains parameter documentation\n")
		}

		// Check for response section
		if page.MustHas("text=Response") || page.MustHas("text=Example Response") {
			fmt.Printf("  ✅ Contains response documentation\n")
		}

		// Try to extract some content
		selectors := []string{"main", "article", "[role='main']", ".content"}
		var text string
		for _, selector := range selectors {
			elem, err := page.Timeout(2 * time.Second).Element(selector)
			if err == nil {
				text, _ = elem.Text()
				break
			}
		}

		if len(text) > 100 {
			fmt.Printf("  ✅ Extracted %d characters\n", len(text))
			fmt.Printf("  Preview: %s...\n", text[:min(200, len(text))])
		} else {
			fmt.Printf("  ⚠️  Only extracted %d characters\n", len(text))
		}

		if !hasMethod {
			fmt.Printf("  ⚠️  No HTTP method found - might be navigation page\n")
		}
	}

	fmt.Println("\n✅ Test complete!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

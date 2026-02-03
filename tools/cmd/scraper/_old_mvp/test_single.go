package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	fmt.Println("🧪 Testing single page scrape...")

	// Launch browser with verbose output
	fmt.Println("1️⃣ Launching browser...")
	launchURL := launcher.New().
		Headless(true).
		MustLaunch()

	fmt.Printf("✅ Browser launched at: %s\n", launchURL)

	browser := rod.New().ControlURL(launchURL).MustConnect()
	defer browser.MustClose()
	fmt.Println("✅ Browser connected")

	page := browser.MustPage()
	defer page.MustClose()
	fmt.Println("✅ Page created")

	// Test URL
	testURL := "https://hyperping.com/docs/api/overview"
	fmt.Printf("\n2️⃣ Navigating to: %s\n", testURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := page.Context(ctx).Navigate(testURL)
	if err != nil {
		log.Fatalf("❌ Navigation failed: %v", err)
	}
	fmt.Println("✅ Navigation complete")

	fmt.Println("\n3️⃣ Waiting for page load...")
	err = page.Context(ctx).WaitLoad()
	if err != nil {
		log.Fatalf("❌ Wait load failed: %v", err)
	}
	fmt.Println("✅ Page loaded")

	fmt.Println("\n4️⃣ Extracting content...")

	// Try to get title
	titleElem, err := page.Element("title")
	if err != nil {
		fmt.Printf("⚠️  No title element: %v\n", err)
	} else {
		title, _ := titleElem.Text()
		fmt.Printf("📄 Title: %s\n", title)
	}

	// Try main selectors
	selectors := []string{"main", "article", ".content", "body"}
	for _, selector := range selectors {
		fmt.Printf("   Trying selector: %s...", selector)
		elem, err := page.Timeout(5 * time.Second).Element(selector)
		if err != nil {
			fmt.Printf(" ❌ Not found\n")
			continue
		}

		text, _ := elem.Text()
		fmt.Printf(" ✅ Found (%d chars)\n", len(text))

		if len(text) > 0 {
			fmt.Printf("\n📝 First 200 chars:\n%s...\n", text[:min(200, len(text))])
			break
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

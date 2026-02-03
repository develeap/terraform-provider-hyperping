package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

// DiscoverURLs finds all API documentation URLs by traversing the navigation menu
// Uses a two-stage approach:
// Stage 1: Find parent pages from main navigation
// Stage 2: Visit each parent page to find child method pages
func DiscoverURLs(ctx context.Context, browser *rod.Browser, baseURL string) ([]DiscoveredURL, error) {
	// Check for cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	log.Println("🔍 Stage 1: Discovering parent API sections...")

	// Stage 1: Get parent pages from main navigation
	parentPages, err := discoverParentPages(browser, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to discover parent pages: %w", err)
	}

	log.Printf("   Found %d parent sections\n", len(parentPages))

	// Stage 2: Visit each parent page to find child methods
	log.Println("\n🔍 Stage 2: Discovering child method pages...")

	allPages := []DiscoveredURL{}
	seenURLs := make(map[string]bool)

	// Add all parent pages first
	for _, parent := range parentPages {
		if !seenURLs[parent.URL] {
			allPages = append(allPages, parent)
			seenURLs[parent.URL] = true
		}
	}

	// Now discover children for each parent
	page := browser.MustPage()
	defer page.MustClose()

	for i, parent := range parentPages {
		// Skip non-API parent pages
		if parent.Section == "overview" {
			continue
		}

		log.Printf("   [%d/%d] Checking %s for child pages...\n", i+1, len(parentPages), parent.Section)

		childPages, err := discoverChildPages(page, parent)
		if err != nil {
			log.Printf("      ⚠️  Failed to discover children: %v\n", err)
			continue
		}

		log.Printf("      Found %d child pages\n", len(childPages))

		// Add children
		for _, child := range childPages {
			if !seenURLs[child.URL] {
				allPages = append(allPages, child)
				seenURLs[child.URL] = true
			}
		}
	}

	log.Printf("\n✅ Discovery complete: %d total URLs\n", len(allPages))

	// Log breakdown
	parents := 0
	children := 0
	for _, d := range allPages {
		if d.IsParent {
			parents++
		} else {
			children++
		}
	}
	log.Printf("   Parents: %d, Children: %d\n", parents, children)

	return allPages, nil
}

// discoverParentPages finds top-level API section pages from main navigation
func discoverParentPages(browser *rod.Browser, baseURL string) ([]DiscoveredURL, error) {
	page := browser.MustPage()
	defer page.MustClose()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Navigate to main API docs page
	if err := page.Context(ctx).Navigate(baseURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", baseURL, err)
	}

	if err := page.Context(ctx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("page load timeout: %w", err)
	}

	time.Sleep(1 * time.Second) // Wait for dynamic content

	var discovered []DiscoveredURL
	seenURLs := make(map[string]bool)

	// Find all links in the sidebar/navigation
	links, err := page.Elements("a[href^='/docs/api/']")
	if err != nil {
		return nil, fmt.Errorf("failed to find navigation links: %w", err)
	}

	for _, link := range links {
		href, err := link.Attribute("href")
		if err != nil || href == nil {
			continue
		}

		url := *href
		if !strings.HasPrefix(url, "http") {
			url = "https://hyperping.com" + url
		}

		// Skip if already seen
		if seenURLs[url] {
			continue
		}
		seenURLs[url] = true

		// Parse section from URL
		parts := strings.Split(strings.TrimPrefix(url, "https://hyperping.com/docs/api/"), "/")
		if len(parts) == 0 {
			continue
		}

		section := parts[0]

		// Only add if it's a top-level section (no method in URL)
		if len(parts) == 1 || (len(parts) == 2 && parts[1] == "") {
			discovered = append(discovered, DiscoveredURL{
				URL:      url,
				Section:  section,
				Method:   "",
				IsParent: true,
			})
		}
	}

	return discovered, nil
}

// discoverChildPages visits a parent page and finds all child method pages
func discoverChildPages(page *rod.Page, parent DiscoveredURL) ([]DiscoveredURL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Navigate to parent page
	if err := page.Context(ctx).Navigate(parent.URL); err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}

	if err := page.Context(ctx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("page load timeout: %w", err)
	}

	time.Sleep(1 * time.Second) // Wait for dynamic content

	var children []DiscoveredURL
	seenURLs := make(map[string]bool)

	// Strategy: Find all links on this page that go to child methods
	// Look for links like /docs/api/{section}/{method}
	links, err := page.Elements("a[href^='/docs/api/" + parent.Section + "/']")
	if err != nil {
		return children, nil // Not an error, just no children
	}

	for _, link := range links {
		href, err := link.Attribute("href")
		if err != nil || href == nil {
			continue
		}

		url := *href
		if !strings.HasPrefix(url, "http") {
			url = "https://hyperping.com" + url
		}

		// Skip if already seen
		if seenURLs[url] {
			continue
		}
		seenURLs[url] = true

		// Parse method from URL
		parts := strings.Split(strings.TrimPrefix(url, "https://hyperping.com/docs/api/"), "/")
		if len(parts) < 2 {
			continue
		}

		section := parts[0]
		method := parts[1]

		// Skip if it's the parent page itself
		if method == "" {
			continue
		}

		children = append(children, DiscoveredURL{
			URL:      url,
			Section:  section,
			Method:   method,
			IsParent: false,
		})
	}

	return children, nil
}

// FilterNewURLs returns only URLs that don't exist in cache or have changed
func FilterNewURLs(discovered []DiscoveredURL, cache Cache) []DiscoveredURL {
	var newURLs []DiscoveredURL

	for _, d := range discovered {
		// Generate filename from URL
		filename := URLToFilename(d.URL)

		// Check if in cache
		if _, exists := cache.Entries[filename]; !exists {
			newURLs = append(newURLs, d)
		}
	}

	return newURLs
}

// URLToFilename converts a URL to a safe filename
// Example: https://hyperping.com/docs/api/monitors/create → monitors_create.json
func URLToFilename(url string) string {
	// Remove protocol and domain
	path := strings.TrimPrefix(url, "https://hyperping.com/docs/api/")
	path = strings.TrimPrefix(path, "https://hyperping.notion.site/")

	// Replace slashes with underscores
	filename := strings.ReplaceAll(path, "/", "_")

	// Handle empty path (root page)
	if filename == "" {
		filename = "index"
	}

	// Clean up special characters
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "&", "_")
	filename = strings.ReplaceAll(filename, "=", "_")

	// Add .json extension
	return filename + ".json"
}

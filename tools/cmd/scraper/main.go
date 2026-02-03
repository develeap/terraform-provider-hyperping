package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
	"github.com/go-rod/rod"
	"golang.org/x/time/rate"
)

func main() {
	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle signals in background
	go func() {
		sig := <-sigChan
		log.Printf("\n🛑 Received signal %v, initiating graceful shutdown...\n", sig)
		cancel()
	}()

	// Run scraper with context
	exitCode := run(ctx)
	os.Exit(exitCode)
}

func run(ctx context.Context) int {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	fmt.Println("🚀 Hyperping API Documentation Scraper - Plan A")
	fmt.Println(strings.Repeat("=", 60))

	// Load configuration
	config := DefaultConfig()
	log.Printf("📋 Configuration:\n")
	log.Printf("   Base URL: %s\n", config.BaseURL)
	log.Printf("   Output Dir: %s\n", config.OutputDir)
	log.Printf("   Rate Limit: %.1f req/sec\n", config.RateLimit)
	log.Printf("   Timeout: %v\n", config.Timeout)
	log.Printf("   Retries: %d\n", config.Retries)
	os.Stdout.Sync()

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Printf("❌ Failed to create output directory: %v\n", err)
		return 1
	}

	// Load cache
	cache, err := LoadCache(config.CacheFile)
	if err != nil {
		log.Printf("❌ Failed to load cache: %v\n", err)
		return 1
	}

	// Launch browser
	log.Println("\n🌐 Launching browser...")

	browser, cleanup, err := launchBrowser(config)
	if err != nil {
		log.Printf("❌ Failed to launch browser: %v\n", err)
		return 1
	}
	defer cleanup() // Always cleanup browser

	log.Println("✅ Browser connected")

	// Phase 1: URL Discovery
	log.Println("\n" + strings.Repeat("=", 60))
	discovered, err := DiscoverURLs(ctx, browser, config.BaseURL)
	if err != nil {
		if ctx.Err() != nil {
			log.Println("❌ Discovery interrupted by shutdown")
			return 2
		}
		log.Printf("❌ URL discovery failed: %v\n", err)
		return 1
	}

	log.Printf("\n📊 Discovery Summary:\n")
	log.Printf("   Total URLs: %d\n", len(discovered))

	// Group by section
	sections := make(map[string]int)
	for _, d := range discovered {
		sections[d.Section]++
	}
	log.Println("   By section:")
	for section, count := range sections {
		log.Printf("      %s: %d\n", section, count)
	}
	os.Stdout.Sync()

	// Phase 2: Smart Scraping with Cache
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🔄 Starting smart scraping...")
	os.Stdout.Sync()

	// Create reusable page
	page := browser.MustPage()
	defer page.MustClose()

	// Rate limiter
	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), 1)

	startTime := time.Now()
	scraped := 0
	skipped := 0
	failed := 0

	newCache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	for i, discoveredURL := range discovered {
		// Check for cancellation
		if ctx.Err() != nil {
			log.Println("\n⚠️  Scraping interrupted by shutdown")
			break
		}

		// Rate limiting with context
		if err := limiter.Wait(ctx); err != nil {
			if ctx.Err() != nil {
				log.Println("\n⚠️  Rate limiter interrupted by shutdown")
				break
			}
			log.Printf("⚠️  Rate limit error: %v\n", err)
			continue
		}

		filename := URLToFilename(discoveredURL.URL)
		log.Printf("\n[%d/%d] %s\n", i+1, len(discovered), discoveredURL.URL)
		log.Printf("        Section: %s, Method: %s\n", discoveredURL.Section, discoveredURL.Method)
		os.Stdout.Sync()

		// Try to scrape with retries (pass context for cancellation)
		pageData, err := scrapeWithRetry(ctx, page, discoveredURL.URL, config.Retries, config.Timeout)
		if err != nil {
			log.Printf("        ❌ Failed after %d retries: %v\n", config.Retries, err)
			failed++
			os.Stdout.Sync()
			continue
		}

		// Check if content changed
		changed := HasChanged(cache, filename, pageData.Text)

		if changed {
			log.Printf("        ✅ Scraped (%d chars) - CHANGED\n", len(pageData.Text))
			scraped++

			// Save to disk
			outputPath := filepath.Join(config.OutputDir, filename)
			if err := savePageData(outputPath, pageData); err != nil {
				log.Printf("        ⚠️  Failed to save: %v\n", err)
			}
		} else {
			log.Printf("        ⏭️  Unchanged (%d chars) - SKIPPED\n", len(pageData.Text))
			skipped++
		}

		// Update cache
		UpdateCache(&newCache, filename, pageData)
		os.Stdout.Sync()
	}

	duration := time.Since(startTime)

	// Phase 3: Results and Cache Update
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📊 Scraping Complete!")
	log.Printf("\n   Total URLs: %d\n", len(discovered))
	log.Printf("   Scraped: %d (content changed)\n", scraped)
	log.Printf("   Skipped: %d (unchanged)\n", skipped)
	log.Printf("   Failed: %d\n", failed)
	log.Printf("   Duration: %v\n", duration.Round(time.Second))

	if skipped > 0 {
		timeSavings := float64(skipped) / float64(len(discovered)) * 100
		log.Printf("   Time Savings: ~%.0f%%\n", timeSavings)
	}

	// Save updated cache
	if err := SaveCache(config.CacheFile, newCache); err != nil {
		log.Printf("⚠️  Failed to save cache: %v\n", err)
	}

	// Compare with old cache
	if len(cache.Entries) > 0 {
		log.Println("\n📈 Change Detection:")
		stats := CompareCaches(cache, newCache)
		log.Printf("   Unchanged: %d\n", stats["unchanged"])
		log.Printf("   Modified: %d\n", stats["modified"])
		log.Printf("   Added: %d\n", stats["added"])
		log.Printf("   Deleted: %d\n", stats["deleted"])
	}

	// Phase 4: Snapshot Management
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📸 Managing API snapshots...")
	os.Stdout.Sync()

	snapshotMgr := NewSnapshotManager("snapshots")

	// Get previous snapshot for comparison
	var previousSnapshot string
	previousSnapshot, err = snapshotMgr.GetLatestSnapshot()
	if err != nil {
		log.Printf("   No previous snapshot found (first run)\n")
	} else {
		log.Printf("   Found previous snapshot: %s\n", filepath.Base(previousSnapshot))
	}

	// Save current snapshot
	currentPages := make(map[string]*extractor.PageData)
	for filename := range newCache.Entries {
		filePath := filepath.Join(config.OutputDir, filename)
		if utils.FileExists(filePath) {
			pageData, err := loadPageData(filePath)
			if err == nil {
				currentPages[filename] = pageData
			}
		}
	}

	currentTimestamp := time.Now()
	if err := snapshotMgr.SaveSnapshot(currentTimestamp, currentPages); err != nil {
		log.Printf("⚠️  Failed to save snapshot: %v\n", err)
	}

	// Phase 5: Semantic Diff Analysis
	var diffs []APIDiff

	if previousSnapshot != "" {
		log.Println("\n" + strings.Repeat("=", 60))
		log.Println("🔍 Analyzing API changes (semantic diff)...")
		os.Stdout.Sync()

		currentSnapshot := filepath.Join("snapshots", currentTimestamp.Format("2006-01-02_15-04-05"))

		diffs, err = snapshotMgr.CompareSnapshots(previousSnapshot, currentSnapshot)
		if err != nil {
			log.Printf("⚠️  Failed to compare snapshots: %v\n", err)
		}
	}

	// Phase 6: GitHub Issue Creation
	if len(diffs) > 0 {
		log.Println("\n" + strings.Repeat("=", 60))
		log.Println("📝 Generating diff report...")
		os.Stdout.Sync()

		report := GenerateDiffReport(diffs, currentTimestamp)

		// Save markdown report locally
		reportFile := "api_changes_" + currentTimestamp.Format("2006-01-02_15-04-05") + ".md"
		if err := SaveDiffReport(report, reportFile); err != nil {
			log.Printf("⚠️  Failed to save diff report: %v\n", err)
		} else {
			log.Printf("✅ Diff report saved: %s\n", reportFile)
		}

		log.Printf("\n📊 Summary: %s\n", report.Summary)
		if report.Breaking {
			log.Printf("   ⚠️  WARNING: Contains breaking changes!\n")
		}

		// Create GitHub issue if configured (or preview mode)
		log.Println("\n🐙 GitHub Integration...")
		githubClient, err := LoadGitHubConfig()
		if err != nil {
			log.Printf("   ⏭️  GitHub credentials not found\n")
			log.Println("   Running in PREVIEW MODE...")
			PreviewGitHubIssue(report, "")
		} else {
			// Ensure labels exist
			if err := githubClient.CreateLabelsIfNeeded(); err != nil {
				log.Printf("   ⚠️  Failed to create labels: %v\n", err)
			}

			// Create issue
			snapshotURL := "" // TODO: Generate URL to snapshot in GitHub repo
			if err := githubClient.CreateIssue(report, snapshotURL); err != nil {
				log.Printf("   ❌ Failed to create issue: %v\n", err)
			}
		}
	} else {
		log.Println("\n✅ No API changes detected")
	}

	// Cleanup old snapshots (keep last 10)
	if err := snapshotMgr.CleanupOldSnapshots(10); err != nil {
		log.Printf("\n⚠️  Failed to cleanup old snapshots: %v\n", err)
	}

	log.Println("\n✅ Done!")
	return 0 // Success
}

// loadPageData loads a PageData struct from a JSON file
func loadPageData(filepath string) (*extractor.PageData, error) {
	var pageData extractor.PageData
	if err := utils.LoadJSON(filepath, &pageData); err != nil {
		return nil, err
	}
	return &pageData, nil
}

// fileExists checks if a file exists
func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

// scrapeWithRetry attempts to scrape a page with exponential backoff
func scrapeWithRetry(ctx context.Context, page *rod.Page, url string, maxRetries int, timeout time.Duration) (*extractor.PageData, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check for cancellation before retry
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if attempt > 0 {
			// Exponential backoff with cancellation support
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			log.Printf("        ⏳ Retry %d/%d after %v...\n", attempt+1, maxRetries, backoff)

			select {
			case <-time.After(backoff):
				// Backoff completed
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		pageData, err := scrapePage(ctx, page, url, timeout)
		if err == nil {
			return pageData, nil
		}

		lastErr = err
		log.Printf("        ⚠️  Attempt %d failed: %v\n", attempt+1, err)
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

// scrapePage scrapes a single page and extracts content
func scrapePage(ctx context.Context, page *rod.Page, url string, timeout time.Duration) (*extractor.PageData, error) {
	// Create timeout context from parent context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Navigate
	if err := page.Context(timeoutCtx).Navigate(url); err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}

	// Wait for page load
	if err := page.Context(timeoutCtx).WaitLoad(); err != nil {
		return nil, fmt.Errorf("page load timeout: %w", err)
	}

	// Wait for content with cancellation support (TODO: replace with proper WaitStable in Phase 3)
	select {
	case <-time.After(500 * time.Millisecond):
		// Wait completed
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Clean up page (remove scripts, styles)
	page.Eval(`() => {
		document.querySelectorAll('script, style, noscript').forEach(el => el.remove());
	}`)

	// Extract title
	title := ""
	if titleElem, err := page.Timeout(5 * time.Second).Element("title"); err == nil {
		title, _ = titleElem.Text()
	}

	// Extract body text
	text := ""
	if bodyElem, err := page.Timeout(5 * time.Second).Element("body"); err == nil {
		text, _ = bodyElem.Text()
	}

	// Get HTML
	html, err := page.HTML()
	if err != nil {
		return nil, fmt.Errorf("failed to get HTML: %w", err)
	}

	return &extractor.PageData{
		URL:       url,
		Title:     strings.TrimSpace(title),
		Text:      strings.TrimSpace(text),
		HTML:      html,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// savePageData writes a PageData struct to a JSON file
func savePageData(filepath string, data *extractor.PageData) error {
	return utils.SaveJSON(filepath, data)
}

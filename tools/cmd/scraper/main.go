package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/analyzer"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/contract"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
	"github.com/go-rod/rod"
	"golang.org/x/time/rate"
)

// Command line flags
var (
	analyzeMode bool
	providerDir string
	snapshotDir string
	cassetteDir string
)

func main() {
	// Parse command line flags
	flag.BoolVar(&analyzeMode, "analyze", false, "Run coverage analysis instead of scraping")
	flag.StringVar(&providerDir, "provider-dir", "../../../internal", "Path to provider internal directory")
	flag.StringVar(&snapshotDir, "snapshot-dir", "./snapshots", "Path to snapshots directory")
	flag.StringVar(&cassetteDir, "cassette-dir", "", "Path to VCR cassettes for contract testing (optional)")
	flag.Parse()

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

	// Run appropriate mode
	var exitCode int
	if analyzeMode {
		exitCode = runAnalyze(ctx)
	} else {
		exitCode = run(ctx)
	}
	os.Exit(exitCode)
}

func run(ctx context.Context) int {
	// Setup configuration
	config, cache, err := setupScraper()
	if err != nil {
		log.Printf("❌ Setup failed: %v\n", err)
		return 1
	}

	// Launch browser
	browser, cleanup, err := launchBrowser(config)
	if err != nil {
		log.Printf("❌ Failed to launch browser: %v\n", err)
		return 1
	}
	defer cleanup()
	log.Println("✅ Browser connected")

	// Discover URLs
	discovered, err := discoverAndSummarize(ctx, browser, config.BaseURL)
	if err != nil {
		if ctx.Err() != nil {
			log.Println("❌ Discovery interrupted by shutdown")
			return 2
		}
		log.Printf("❌ URL discovery failed: %v\n", err)
		return 1
	}

	// Scrape pages
	newCache, stats, err := scrapePages(ctx, browser, discovered, cache, config)
	if err != nil {
		if ctx.Err() != nil {
			log.Println("❌ Scraping interrupted by shutdown")
			return 2
		}
		log.Printf("❌ Scraping failed: %v\n", err)
		return 1
	}

	// Report results
	if err := reportResults(discovered, stats, cache, newCache, config.CacheFile); err != nil {
		log.Printf("⚠️  Failed to save results: %v\n", err)
	}

	// Manage snapshots and analyze changes
	if err := manageSnapshots(ctx, config.OutputDir, newCache); err != nil {
		log.Printf("⚠️  Snapshot management failed: %v\n", err)
	}

	log.Println("\n✅ Done!")
	return 0
}

// runAnalyze performs coverage analysis comparing API docs to provider schema
func runAnalyze(ctx context.Context) int {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	fmt.Println("🔍 Hyperping Provider Coverage Analyzer")
	fmt.Println(strings.Repeat("=", 60))

	log.Printf("📋 Configuration:\n")
	log.Printf("   Provider Dir: %s\n", providerDir)
	log.Printf("   Snapshot Dir: %s\n", snapshotDir)
	if cassetteDir != "" {
		log.Printf("   Cassette Dir: %s\n", cassetteDir)
	}

	// Find latest snapshot
	snapshotMgr := NewSnapshotManager(snapshotDir)
	latestSnapshot, err := snapshotMgr.GetLatestSnapshot()
	if err != nil {
		log.Printf("❌ No snapshots found. Run scraper first: %v\n", err)
		return 1
	}
	log.Printf("   Latest Snapshot: %s\n", filepath.Base(latestSnapshot))

	// Initialize analyzer
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📦 Extracting provider schemas...")

	a, err := analyzer.NewAnalyzer(providerDir)
	if err != nil {
		log.Printf("❌ Failed to initialize analyzer: %v\n", err)
		return 1
	}
	log.Printf("   Found %d resource schemas\n", len(a.ProviderSchemas))

	for _, schema := range a.ProviderSchemas {
		log.Printf("   - %s (%d fields)\n", schema.Name, len(schema.Fields))
	}

	// Load API parameters from snapshot
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📄 Loading API parameters from snapshot...")

	apiParams, err := analyzer.LoadAPIParamsFromSnapshot(latestSnapshot)
	if err != nil {
		log.Printf("❌ Failed to load API params: %v\n", err)
		return 1
	}
	log.Printf("   Loaded %d endpoint files\n", len(apiParams))

	// Run coverage analysis
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🔬 Analyzing coverage...")

	report := a.AnalyzeCoverage(apiParams)

	// Print summary
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📊 Coverage Report")
	log.Printf("\n   Overall Coverage: %.1f%%\n", report.CoveragePercent)
	log.Printf("   Total API Fields: %d\n", report.TotalAPIFields)
	log.Printf("   Covered Fields: %d\n", report.CoveredFields)
	log.Printf("   Missing Fields: %d\n", report.MissingFields)
	log.Printf("   Stale Fields: %d\n", report.StaleFields)

	log.Println("\n   By Resource:")
	for _, rc := range report.Resources {
		log.Printf("   - %s: %.1f%% (%d/%d)\n",
			rc.TerraformResource, rc.CoveragePercent, rc.ImplementedFields, rc.APIFields)
	}

	// Save markdown report
	reportFile := "coverage_report_" + time.Now().Format("2006-01-02_15-04-05") + ".md"
	markdown := analyzer.FormatCoverageReportMarkdown(report)
	if err := utils.SaveToFile(reportFile, markdown); err != nil {
		log.Printf("⚠️  Failed to save report: %v\n", err)
	} else {
		log.Printf("\n✅ Report saved: %s\n", reportFile)
	}

	// Endpoint version analysis
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🔗 Endpoint Version Analysis...")

	if err := runEndpointVersionAnalysis(latestSnapshot, providerDir); err != nil {
		log.Printf("⚠️  Endpoint version analysis failed: %v\n", err)
	}

	// Contract testing analysis (if cassettes provided)
	if cassetteDir != "" {
		if err := runContractAnalysis(cassetteDir, apiParams); err != nil {
			log.Printf("⚠️  Contract analysis failed: %v\n", err)
		}
	}

	// Create GitHub issues for resources with gaps
	if len(report.Gaps) > 0 {
		log.Println("\n" + strings.Repeat("=", 60))
		log.Println("🐙 GitHub Integration...")

		if err := createCoverageIssues(report); err != nil {
			log.Printf("⚠️  Failed to create GitHub issues: %v\n", err)
		}
	}

	log.Println("\n✅ Analysis complete!")
	return 0
}

// runContractAnalysis performs contract testing analysis using VCR cassettes
func runContractAnalysis(cassetteDir string, apiParams map[string][]extractor.APIParameter) error {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🔬 Contract Testing Analysis...")
	log.Printf("   Cassette Dir: %s\n", cassetteDir)

	// Extract schema from cassettes
	cassetteSchema, err := contract.ExtractFromCassettes(cassetteDir)
	if err != nil {
		return fmt.Errorf("failed to extract cassette schema: %w", err)
	}
	log.Printf("   Found %d endpoints in cassettes\n", len(cassetteSchema.Endpoints))

	// Convert API params to documented fields format
	docFields := convertAPIParamsToDocFields(apiParams)

	// Compare cassettes with documentation
	results := contract.CompareWithDocumentation(cassetteSchema, docFields)

	// Count discoveries
	totalUndocumented := 0
	totalMismatches := 0
	for _, r := range results {
		totalUndocumented += r.Summary.UndocumentedFields
		totalMismatches += r.Summary.TypeMismatches
	}

	log.Printf("\n   📊 Contract Testing Results:\n")
	log.Printf("   Undocumented Fields: %d\n", totalUndocumented)
	log.Printf("   Type Mismatches: %d\n", totalMismatches)

	// Generate and save discovery report
	if totalUndocumented > 0 || totalMismatches > 0 {
		discoveryReport := contract.GenerateDiscoveryReport(results)
		reportFile := "discovery_report_" + time.Now().Format("2006-01-02_15-04-05") + ".md"
		if err := utils.SaveToFile(reportFile, discoveryReport); err != nil {
			log.Printf("⚠️  Failed to save discovery report: %v\n", err)
		} else {
			log.Printf("   ✅ Discovery report saved: %s\n", reportFile)
		}
	} else {
		log.Println("   ✅ No documentation gaps discovered via contract testing")
	}

	return nil
}

// runEndpointVersionAnalysis compares API endpoint versions between docs and provider code
func runEndpointVersionAnalysis(snapshotDir, providerDir string) error {
	// Extract endpoints from scraped docs
	docsEndpoints, err := analyzer.ExtractEndpointsFromDocs(snapshotDir)
	if err != nil {
		return fmt.Errorf("failed to extract docs endpoints: %w", err)
	}
	log.Printf("   Found %d endpoints in docs\n", len(docsEndpoints))

	// Extract endpoints from provider code
	providerEndpoints, err := analyzer.ExtractEndpointsFromProvider(providerDir)
	if err != nil {
		return fmt.Errorf("failed to extract provider endpoints: %w", err)
	}
	log.Printf("   Found %d endpoints in provider\n", len(providerEndpoints))

	// Compare versions
	report := analyzer.CompareEndpointVersions(docsEndpoints, providerEndpoints)

	if len(report.Mismatches) > 0 {
		log.Printf("\n   ⚠️  Found %d endpoint version mismatches:\n", len(report.Mismatches))
		for _, m := range report.Mismatches {
			log.Printf("   - %s: docs=%s, provider=%s\n", m.Resource, m.DocsVersion, m.ProviderVersion)
			log.Printf("     Fix: %s\n", m.Suggestion)
		}

		// Save detailed report
		reportContent := analyzer.FormatEndpointReport(report)
		reportFile := "endpoint_version_report_" + time.Now().Format("2006-01-02_15-04-05") + ".md"
		if err := utils.SaveToFile(reportFile, reportContent); err != nil {
			log.Printf("⚠️  Failed to save endpoint report: %v\n", err)
		} else {
			log.Printf("   ✅ Endpoint report saved: %s\n", reportFile)
		}
	} else {
		log.Println("   ✅ All endpoint versions match!")
	}

	return nil
}

// convertAPIParamsToDocFields converts scraped API params to the format expected by contract comparator
func convertAPIParamsToDocFields(apiParams map[string][]extractor.APIParameter) map[string][]contract.DocumentedField {
	result := make(map[string][]contract.DocumentedField)

	for endpoint, params := range apiParams {
		// Extract resource name from endpoint (e.g., "healthchecks_create" -> "healthchecks")
		resource := strings.Split(endpoint, "_")[0]

		for _, param := range params {
			field := contract.DocumentedField{
				Name:     param.Name,
				Type:     param.Type,
				Required: param.Required,
			}
			result[resource] = append(result[resource], field)
		}
	}

	return result
}

// createCoverageIssues creates GitHub issues for coverage gaps
func createCoverageIssues(report *analyzer.CoverageReport) error {
	githubClient, err := LoadGitHubConfig()
	if err != nil {
		log.Printf("   ⏭️  GitHub credentials not found - skipping issue creation\n")
		return nil
	}

	// Group gaps by resource
	gapsByResource := analyzer.GroupGapsByResource(report.Gaps)

	for resource, gaps := range gapsByResource {
		// Get coverage stats for this resource
		var resourceCoverage *analyzer.ResourceCoverage
		for i := range report.Resources {
			if report.Resources[i].Resource == resource {
				resourceCoverage = &report.Resources[i]
				break
			}
		}

		if resourceCoverage == nil {
			continue
		}

		// Count missing fields
		missingCount := 0
		for _, gap := range gaps {
			if gap.Type == analyzer.GapMissing {
				missingCount++
			}
		}

		if missingCount == 0 {
			continue
		}

		// Generate issue content
		title := analyzer.FormatGitHubIssueTitle(resourceCoverage.TerraformResource, missingCount)
		body := analyzer.FormatGitHubIssueBody(resource, resourceCoverage.TerraformResource, *resourceCoverage, gaps)
		labels := analyzer.GetGitHubIssueLabels()

		log.Printf("   Creating issue for %s (%d gaps)...\n", resourceCoverage.TerraformResource, missingCount)

		// Check if issue already exists and update, or create new
		if err := githubClient.CreateOrUpdateCoverageIssue(title, body, labels); err != nil {
			log.Printf("   ⚠️  Failed to create issue for %s: %v\n", resource, err)
		} else {
			log.Printf("   ✅ Issue created/updated for %s\n", resource)
		}
	}

	return nil
}

// setupScraper initializes configuration and loads cache
func setupScraper() (ScraperConfig, Cache, error) {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	fmt.Println("🚀 Hyperping API Documentation Scraper - Plan A")
	fmt.Println(strings.Repeat("=", 60))

	config := DefaultConfig()
	log.Printf("📋 Configuration:\n")
	log.Printf("   Base URL: %s\n", config.BaseURL)
	log.Printf("   Output Dir: %s\n", config.OutputDir)
	log.Printf("   Rate Limit: %.1f req/sec\n", config.RateLimit)
	log.Printf("   Timeout: %v\n", config.Timeout)
	log.Printf("   Retries: %d\n", config.Retries)

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0750); err != nil {
		return config, Cache{}, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load cache
	cache, err := LoadCache(config.CacheFile)
	if err != nil {
		return config, Cache{}, fmt.Errorf("failed to load cache: %w", err)
	}

	return config, cache, nil
}

// discoverAndSummarize discovers URLs and prints summary statistics
func discoverAndSummarize(ctx context.Context, browser *rod.Browser, baseURL string) ([]DiscoveredURL, error) {
	log.Println("\n" + strings.Repeat("=", 60))
	discovered, err := DiscoverURLs(ctx, browser, baseURL)
	if err != nil {
		return nil, err
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

	return discovered, nil
}

// ScrapeStats holds scraping statistics
type ScrapeStats struct {
	Scraped  int
	Skipped  int
	Failed   int
	Duration time.Duration
}

// scrapePages performs the main scraping loop with rate limiting
func scrapePages(ctx context.Context, browser *rod.Browser, discovered []DiscoveredURL, cache Cache, config ScraperConfig) (Cache, ScrapeStats, error) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🔄 Starting smart scraping...")

	page := browser.MustPage()
	defer page.MustClose()

	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), 1)
	startTime := time.Now()

	newCache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	stats := ScrapeStats{}

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

		// Scrape single URL
		if err := scrapeSingleURL(ctx, page, discoveredURL, i, len(discovered), cache, &newCache, &stats, config); err != nil {
			if ctx.Err() != nil {
				break
			}
			// Error already logged in scrapeSingleURL
		}
	}

	stats.Duration = time.Since(startTime)
	return newCache, stats, nil
}

// scrapeSingleURL scrapes a single URL and updates statistics
func scrapeSingleURL(ctx context.Context, page *rod.Page, discoveredURL DiscoveredURL, index, total int, oldCache Cache, newCache *Cache, stats *ScrapeStats, config ScraperConfig) error {
	filename := URLToFilename(discoveredURL.URL)
	log.Printf("\n[%d/%d] %s\n", index+1, total, discoveredURL.URL)
	log.Printf("        Section: %s, Method: %s\n", discoveredURL.Section, discoveredURL.Method)

	// Try to scrape with retries
	pageData, err := scrapeWithRetry(ctx, page, discoveredURL.URL, config.Retries, config.Timeout)
	if err != nil {
		log.Printf("        ❌ Failed after %d retries: %v\n", config.Retries, err)
		stats.Failed++
		return err
	}

	// Check if content changed
	changed := HasChanged(oldCache, filename, pageData.Text)

	if changed {
		log.Printf("        ✅ Scraped (%d chars) - CHANGED\n", len(pageData.Text))
		stats.Scraped++

		// Save to disk
		outputPath := filepath.Join(config.OutputDir, filename)
		if err := savePageData(outputPath, pageData); err != nil {
			log.Printf("        ⚠️  Failed to save: %v\n", err)
		}
	} else {
		log.Printf("        ⏭️  Unchanged (%d chars) - SKIPPED\n", len(pageData.Text))
		stats.Skipped++
	}

	// Update cache
	UpdateCache(newCache, filename, pageData)
	return nil
}

// reportResults prints scraping statistics and saves cache
func reportResults(discovered []DiscoveredURL, stats ScrapeStats, oldCache, newCache Cache, cacheFile string) error {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📊 Scraping Complete!")
	log.Printf("\n   Total URLs: %d\n", len(discovered))
	log.Printf("   Scraped: %d (content changed)\n", stats.Scraped)
	log.Printf("   Skipped: %d (unchanged)\n", stats.Skipped)
	log.Printf("   Failed: %d\n", stats.Failed)
	log.Printf("   Duration: %v\n", stats.Duration.Round(time.Second))

	if stats.Skipped > 0 {
		timeSavings := float64(stats.Skipped) / float64(len(discovered)) * 100
		log.Printf("   Time Savings: ~%.0f%%\n", timeSavings)
	}

	// Save updated cache
	if err := SaveCache(cacheFile, newCache); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	// Compare with old cache
	if len(oldCache.Entries) > 0 {
		log.Println("\n📈 Change Detection:")
		cacheStats := CompareCaches(oldCache, newCache)
		log.Printf("   Unchanged: %d\n", cacheStats["unchanged"])
		log.Printf("   Modified: %d\n", cacheStats["modified"])
		log.Printf("   Added: %d\n", cacheStats["added"])
		log.Printf("   Deleted: %d\n", cacheStats["deleted"])
	}

	return nil
}

// manageSnapshots handles snapshot creation, comparison, and reporting
func manageSnapshots(ctx context.Context, outputDir string, newCache Cache) error {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📸 Managing API snapshots...")

	snapshotMgr := NewSnapshotManager("snapshots")

	// Get previous snapshot for comparison
	previousSnapshot, err := snapshotMgr.GetLatestSnapshot()
	if err != nil {
		log.Printf("   No previous snapshot found (first run)\n")
	} else {
		log.Printf("   Found previous snapshot: %s\n", filepath.Base(previousSnapshot))
	}

	// Save current snapshot
	currentPages := loadCurrentPages(outputDir, newCache)
	currentTimestamp := time.Now()

	if err := snapshotMgr.SaveSnapshot(currentTimestamp, currentPages); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	// Analyze changes if previous snapshot exists
	if previousSnapshot != "" {
		diffs, err := analyzeChanges(snapshotMgr, previousSnapshot, currentTimestamp)
		if err != nil {
			log.Printf("⚠️  Failed to compare snapshots: %v\n", err)
		} else if len(diffs) > 0 {
			if err := generateAndReportDiffs(diffs, currentTimestamp); err != nil {
				return err
			}
		} else {
			log.Println("\n✅ No API changes detected")
		}
	}

	// Cleanup old snapshots
	if err := snapshotMgr.CleanupOldSnapshots(10); err != nil {
		log.Printf("\n⚠️  Failed to cleanup old snapshots: %v\n", err)
	}

	return nil
}

// loadCurrentPages loads all page data from disk for snapshot
func loadCurrentPages(outputDir string, newCache Cache) map[string]*extractor.PageData {
	currentPages := make(map[string]*extractor.PageData)
	for filename := range newCache.Entries {
		filePath := filepath.Join(outputDir, filename)
		if utils.FileExists(filePath) {
			pageData, err := loadPageData(filePath)
			if err == nil {
				currentPages[filename] = pageData
			}
		}
	}
	return currentPages
}

// analyzeChanges compares snapshots and returns differences
func analyzeChanges(snapshotMgr *SnapshotManager, previousSnapshot string, currentTimestamp time.Time) ([]APIDiff, error) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🔍 Analyzing API changes (semantic diff)...")

	currentSnapshot := filepath.Join("snapshots", currentTimestamp.Format("2006-01-02_15-04-05"))
	return snapshotMgr.CompareSnapshots(previousSnapshot, currentSnapshot)
}

// generateAndReportDiffs creates diff report and GitHub issue
func generateAndReportDiffs(diffs []APIDiff, timestamp time.Time) error {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📝 Generating diff report...")

	report := GenerateDiffReport(diffs, timestamp)

	// Save markdown report locally
	reportFile := "api_changes_" + timestamp.Format("2006-01-02_15-04-05") + ".md"
	if err := SaveDiffReport(report, reportFile); err != nil {
		return fmt.Errorf("failed to save diff report: %w", err)
	}
	log.Printf("✅ Diff report saved: %s\n", reportFile)

	log.Printf("\n📊 Summary: %s\n", report.Summary)
	if report.Breaking {
		log.Printf("   ⚠️  WARNING: Contains breaking changes!\n")
	}

	// Create GitHub issue if configured
	return createGitHubIssue(report)
}

// createGitHubIssue creates a GitHub issue for API changes
func createGitHubIssue(report DiffReport) error {
	log.Println("\n🐙 GitHub Integration...")

	// Skip issue creation if no actual API changes detected
	if report.ChangedPages == 0 {
		log.Println("   ⏭️  No API changes to report - skipping issue creation")
		return nil
	}

	githubClient, err := LoadGitHubConfig()
	if err != nil {
		log.Printf("   ⏭️  GitHub credentials not found\n")
		log.Println("   Running in PREVIEW MODE...")
		PreviewGitHubIssue(report, "")
		return nil
	}

	// Ensure labels exist
	if err := githubClient.CreateLabelsIfNeeded(); err != nil {
		log.Printf("   ⚠️  Failed to create labels: %v\n", err)
	}

	// Create issue
	snapshotURL := "" // TODO: Generate URL to snapshot in GitHub repo
	if err := githubClient.CreateIssue(report, snapshotURL); err != nil {
		return fmt.Errorf("failed to create GitHub issue: %w", err)
	}

	return nil
}

// loadPageData loads a PageData struct from a JSON file
func loadPageData(filepath string) (*extractor.PageData, error) {
	var pageData extractor.PageData
	if err := utils.LoadJSON(filepath, &pageData); err != nil {
		return nil, err
	}
	return &pageData, nil
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	errStr := err.Error()

	// Don't retry client errors (404, 403, 401, etc.)
	nonRetryablePatterns := []string{
		"404",
		"403",
		"401",
		"400",
		"not found",
		"forbidden",
		"unauthorized",
		"bad request",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return false
		}
	}

	// Retry network errors, timeouts, 500s, etc.
	return true
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

		// Check if error is retryable
		if !isRetryableError(err) {
			log.Printf("        ⚠️  Non-retryable error: %v\n", err)
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}

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

	// Wait for DOM to stabilize (no changes for 100ms)
	// This handles dynamic content loading better than fixed sleep
	if err := page.Context(timeoutCtx).WaitStable(100 * time.Millisecond); err != nil {
		// WaitStable timeout is not fatal - page might just be slow
		// Log and continue with whatever content we have
		log.Printf("        ⚠️  Page stability timeout (content may still be loading)\n")
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

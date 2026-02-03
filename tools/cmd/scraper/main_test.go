package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
)

func TestSetupScraper(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	config, cache, err := setupScraper()
	if err != nil {
		t.Fatalf("setupScraper() failed: %v", err)
	}

	// Verify config has defaults
	if config.BaseURL == "" {
		t.Error("Expected non-empty BaseURL")
	}
	if config.RateLimit <= 0 {
		t.Error("Expected positive RateLimit")
	}

	// Verify cache structure
	if cache.Entries == nil {
		t.Error("Expected cache.Entries to be initialized")
	}

	// Verify output directory was created
	if _, err := os.Stat(config.OutputDir); os.IsNotExist(err) {
		t.Errorf("Expected output directory %s to be created", config.OutputDir)
	}
}

func TestScrapeStats(t *testing.T) {
	stats := ScrapeStats{
		Scraped:  10,
		Skipped:  5,
		Failed:   2,
		Duration: 30 * time.Second,
	}

	if stats.Scraped != 10 {
		t.Errorf("Expected Scraped=10, got %d", stats.Scraped)
	}
	if stats.Skipped != 5 {
		t.Errorf("Expected Skipped=5, got %d", stats.Skipped)
	}
	if stats.Failed != 2 {
		t.Errorf("Expected Failed=2, got %d", stats.Failed)
	}
	if stats.Duration != 30*time.Second {
		t.Errorf("Expected Duration=30s, got %v", stats.Duration)
	}
}

func TestLoadPageData(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test page data
	testData := &extractor.PageData{
		URL:       "https://example.com",
		Title:     "Test Page",
		Text:      "Test content",
		HTML:      "<html>test</html>",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Save test data
	testFile := filepath.Join(tmpDir, "test.json")
	if err := savePageData(testFile, testData); err != nil {
		t.Fatalf("Failed to save test data: %v", err)
	}

	// Load and verify
	loaded, err := loadPageData(testFile)
	if err != nil {
		t.Fatalf("loadPageData() failed: %v", err)
	}

	if loaded.URL != testData.URL {
		t.Errorf("Expected URL=%s, got %s", testData.URL, loaded.URL)
	}
	if loaded.Title != testData.Title {
		t.Errorf("Expected Title=%s, got %s", testData.Title, loaded.Title)
	}
	if loaded.Text != testData.Text {
		t.Errorf("Expected Text=%s, got %s", testData.Text, loaded.Text)
	}
}

func TestLoadPageData_FileNotFound(t *testing.T) {
	_, err := loadPageData("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestSavePageData(t *testing.T) {
	tmpDir := t.TempDir()

	testData := &extractor.PageData{
		URL:       "https://example.com/api",
		Title:     "API Documentation",
		Text:      "API content here",
		HTML:      "<html><body>API</body></html>",
		Timestamp: "2026-02-03T10:00:00Z",
	}

	testFile := filepath.Join(tmpDir, "api.json")
	err := savePageData(testFile, testData)
	if err != nil {
		t.Fatalf("savePageData() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", testFile)
	}

	// Load and verify content
	loaded, err := loadPageData(testFile)
	if err != nil {
		t.Fatalf("Failed to load saved data: %v", err)
	}

	if loaded.URL != testData.URL {
		t.Errorf("Expected URL=%s, got %s", testData.URL, loaded.URL)
	}
}

func TestLoadCurrentPages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test cache
	cache := Cache{
		Entries: map[string]CacheEntry{
			"page1.json": {Filename: "page1.json", ContentHash: "hash1", LastModified: time.Now()},
			"page2.json": {Filename: "page2.json", ContentHash: "hash2", LastModified: time.Now()},
		},
		CreatedAt: time.Now(),
	}

	// Create test page files
	page1 := &extractor.PageData{
		URL:       "https://example.com/page1",
		Title:     "Page 1",
		Text:      "Content 1",
		HTML:      "<html>1</html>",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	page2 := &extractor.PageData{
		URL:       "https://example.com/page2",
		Title:     "Page 2",
		Text:      "Content 2",
		HTML:      "<html>2</html>",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	savePageData(filepath.Join(tmpDir, "page1.json"), page1)
	savePageData(filepath.Join(tmpDir, "page2.json"), page2)

	// Load current pages
	pages := loadCurrentPages(tmpDir, cache)

	if len(pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(pages))
	}

	if pages["page1.json"] == nil {
		t.Error("Expected page1.json to be loaded")
	}
	if pages["page2.json"] == nil {
		t.Error("Expected page2.json to be loaded")
	}

	if pages["page1.json"].URL != page1.URL {
		t.Errorf("Expected page1 URL=%s, got %s", page1.URL, pages["page1.json"].URL)
	}
}

func TestLoadCurrentPages_MissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	cache := Cache{
		Entries: map[string]CacheEntry{
			"missing.json": {Filename: "missing.json", ContentHash: "hash", LastModified: time.Now()},
		},
		CreatedAt: time.Now(),
	}

	pages := loadCurrentPages(tmpDir, cache)

	// Should handle missing files gracefully
	if len(pages) != 0 {
		t.Errorf("Expected 0 pages for missing files, got %d", len(pages))
	}
}

func TestReportResults_EmptyCache(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "test_cache.json")

	discovered := []DiscoveredURL{
		{URL: "https://example.com/1", Section: "api", Method: "GET"},
		{URL: "https://example.com/2", Section: "api", Method: "POST"},
	}

	stats := ScrapeStats{
		Scraped:  2,
		Skipped:  0,
		Failed:   0,
		Duration: 10 * time.Second,
	}

	oldCache := Cache{Entries: make(map[string]CacheEntry)}
	newCache := Cache{
		Entries: map[string]CacheEntry{
			"page1.json": {Filename: "page1.json", ContentHash: "hash1", LastModified: time.Now()},
			"page2.json": {Filename: "page2.json", ContentHash: "hash2", LastModified: time.Now()},
		},
		CreatedAt: time.Now(),
	}

	err := reportResults(discovered, stats, oldCache, newCache, cacheFile)
	if err != nil {
		t.Fatalf("reportResults() failed: %v", err)
	}

	// Verify cache file was created
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Errorf("Expected cache file %s to be created", cacheFile)
	}
}

func TestReportResults_WithSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "test_cache.json")

	discovered := []DiscoveredURL{
		{URL: "https://example.com/1", Section: "api", Method: "GET"},
		{URL: "https://example.com/2", Section: "api", Method: "POST"},
		{URL: "https://example.com/3", Section: "api", Method: "PUT"},
	}

	stats := ScrapeStats{
		Scraped:  1,
		Skipped:  2,
		Failed:   0,
		Duration: 15 * time.Second,
	}

	oldCache := Cache{
		Entries: map[string]CacheEntry{
			"page1.json": {Filename: "page1.json", ContentHash: "oldhash1", LastModified: time.Now().Add(-1 * time.Hour)},
			"page2.json": {Filename: "page2.json", ContentHash: "oldhash2", LastModified: time.Now().Add(-1 * time.Hour)},
		},
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	newCache := Cache{
		Entries: map[string]CacheEntry{
			"page1.json": {Filename: "page1.json", ContentHash: "newhash1", LastModified: time.Now()},
			"page2.json": {Filename: "page2.json", ContentHash: "oldhash2", LastModified: time.Now()}, // Same hash
			"page3.json": {Filename: "page3.json", ContentHash: "newhash3", LastModified: time.Now()},
		},
		CreatedAt: time.Now(),
	}

	err := reportResults(discovered, stats, oldCache, newCache, cacheFile)
	if err != nil {
		t.Fatalf("reportResults() failed: %v", err)
	}

	// Should calculate time savings
	// Expected: 2 skipped out of 3 = ~66% savings (logged)
}

func TestRun_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This should return exit code 2 (interrupted)
	// Note: This is more of an integration test
	// We'd need to mock browser for a proper unit test
	// Placeholder for future enhancement - requires browser mock
	t.Skip("Integration test - requires browser mock implementation")
}

func TestDiscoverAndSummarize_ContextCancelled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Note: This requires browser mock for proper testing
	// Would need to refactor to inject browser interface
	// Placeholder for future enhancement
	t.Skip("Integration test - requires browser mock implementation")
}

package main

import (
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
)

func TestDefaultCacheManager(t *testing.T) {
	manager := &DefaultCacheManager{}

	// Test Update
	cache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	pageData := &extractor.PageData{
		URL:       "https://example.com",
		Text:      "test content",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	manager.Update(&cache, "test.json", pageData)

	if len(cache.Entries) != 1 {
		t.Errorf("Expected 1 cache entry, got %d", len(cache.Entries))
	}

	// Test HasChanged
	changed := manager.HasChanged(cache, "test.json", "test content")
	if changed {
		t.Error("Expected no change for same content")
	}

	changed = manager.HasChanged(cache, "test.json", "different content")
	if !changed {
		t.Error("Expected change for different content")
	}
}

func TestDefaultURLDiscoverer(t *testing.T) {
	discoverer := &DefaultURLDiscoverer{}

	// Test ToFilename
	filename := discoverer.ToFilename("https://hyperping.com/docs/api/monitors/create")
	expected := "monitors_create.json"
	if filename != expected {
		t.Errorf("Expected %s, got %s", expected, filename)
	}

	// Test FilterNew
	cache := Cache{
		Entries: map[string]CacheEntry{
			"monitors.json": {Filename: "monitors.json", ContentHash: "hash1", LastModified: time.Now()},
		},
		CreatedAt: time.Now(),
	}

	discovered := []DiscoveredURL{
		{URL: "https://hyperping.com/docs/api/monitors", Section: "monitors"},
		{URL: "https://hyperping.com/docs/api/incidents", Section: "incidents"},
	}

	newURLs := discoverer.FilterNew(discovered, cache)
	if len(newURLs) != 1 {
		t.Errorf("Expected 1 new URL, got %d", len(newURLs))
	}
	if newURLs[0].Section != "incidents" {
		t.Errorf("Expected incidents section, got %s", newURLs[0].Section)
	}
}

func TestDefaultSnapshotStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewDefaultSnapshotStorage(tmpDir)

	// Test Save
	pages := map[string]*extractor.PageData{
		"test.json": {
			URL:       "https://example.com",
			Title:     "Test",
			Text:      "Content",
			HTML:      "<html>test</html>",
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}

	timestamp := time.Now()
	err := storage.Save(timestamp, pages)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test GetLatest
	latest, err := storage.GetLatest()
	if err != nil {
		t.Fatalf("GetLatest failed: %v", err)
	}
	if latest == "" {
		t.Error("Expected latest snapshot path")
	}
}

func TestDefaultDiffReporter(t *testing.T) {
	reporter := &DefaultDiffReporter{}

	diffs := []APIDiff{
		{
			Section:  "monitors",
			Endpoint: "/monitors/create",
			Method:   "POST",
			Breaking: false,
		},
	}

	timestamp := time.Now()
	report := reporter.Generate(diffs, timestamp)

	if report.TotalPages != 0 {
		t.Errorf("Expected TotalPages=0, got %d", report.TotalPages)
	}
	if len(report.APIDiffs) != 1 {
		t.Errorf("Expected 1 diff, got %d", len(report.APIDiffs))
	}
}

func TestDefaultConfigProvider(t *testing.T) {
	provider := NewDefaultConfigProvider()

	config := provider.Get()

	if config.BaseURL == "" {
		t.Error("Expected non-empty BaseURL")
	}
	if config.RateLimit <= 0 {
		t.Error("Expected positive RateLimit")
	}

	err := provider.Validate()
	if err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestLoggerAdapter(t *testing.T) {
	logger := NewLogger(testWriter{t}, LogFormatText)
	adapter := NewLoggerAdapter(logger)

	// Test that methods don't panic
	adapter.Debug("test")
	adapter.Info("test")
	adapter.Warn("test")
	adapter.Error("test")
	adapter.SetLevel(LogLevelDebug)
}

// testWriter is a helper writer for tests
type testWriter struct {
	t *testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestMetricsAdapter(t *testing.T) {
	metrics := NewMetrics()
	adapter := NewMetricsAdapter(metrics)

	adapter.RecordPageScraped(100 * time.Millisecond)
	adapter.RecordPageSkipped()
	adapter.RecordPageFailed()
	adapter.RecordCacheHit()
	adapter.RecordCacheMiss()

	snapshot := adapter.GetSnapshot()
	if snapshot.PagesScraped != 1 {
		t.Errorf("Expected PagesScraped=1, got %d", snapshot.PagesScraped)
	}
	if snapshot.PagesSkipped != 1 {
		t.Errorf("Expected PagesSkipped=1, got %d", snapshot.PagesSkipped)
	}

	summary := adapter.Summary()
	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	prometheus := adapter.PrometheusFormat()
	if prometheus == "" {
		t.Error("Expected non-empty Prometheus format")
	}
}

func TestNewScraper(t *testing.T) {
	config := DefaultConfig()
	cache := &DefaultCacheManager{}
	discoverer := &DefaultURLDiscoverer{}
	scraper := &DefaultPageScraper{}
	snapshots := NewDefaultSnapshotStorage(t.TempDir())
	differ := &DefaultDiffReporter{}
	metrics := NewMetricsAdapter(NewMetrics())
	logger := NewLoggerAdapter(NewLogger(testWriter{t}, LogFormatText))

	// GitHub integration might fail without credentials, so skip it
	var github GitHubIntegration = nil

	s := NewScraper(config, cache, discoverer, scraper, snapshots, differ, github, metrics, logger)

	if s == nil {
		t.Fatal("Expected scraper to be created")
	}
	if s.config.BaseURL == "" {
		t.Error("Expected non-empty config")
	}
}

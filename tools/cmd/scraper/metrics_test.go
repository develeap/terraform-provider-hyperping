package main

import (
	"strings"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()

	if metrics == nil {
		t.Fatal("Expected metrics to be created")
	}
	if metrics.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
}

func TestMetricsRecordPageScraped(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordPageScraped(100 * time.Millisecond)
	metrics.RecordPageScraped(200 * time.Millisecond)

	snapshot := metrics.GetSnapshot()

	if snapshot.PagesScraped != 2 {
		t.Errorf("Expected PagesScraped=2, got %d", snapshot.PagesScraped)
	}
	if snapshot.AvgPageDuration == 0 {
		t.Error("Expected AvgPageDuration to be set")
	}
}

func TestMetricsRecordPageSkipped(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordPageSkipped()
	metrics.RecordPageSkipped()
	metrics.RecordPageSkipped()

	snapshot := metrics.GetSnapshot()

	if snapshot.PagesSkipped != 3 {
		t.Errorf("Expected PagesSkipped=3, got %d", snapshot.PagesSkipped)
	}
}

func TestMetricsRecordPageFailed(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordPageFailed()

	snapshot := metrics.GetSnapshot()

	if snapshot.PagesFailed != 1 {
		t.Errorf("Expected PagesFailed=1, got %d", snapshot.PagesFailed)
	}
}

func TestMetricsCacheOperations(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordCacheHit()
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()
	metrics.SetCacheSize(100)

	snapshot := metrics.GetSnapshot()

	if snapshot.CacheHits != 2 {
		t.Errorf("Expected CacheHits=2, got %d", snapshot.CacheHits)
	}
	if snapshot.CacheMisses != 1 {
		t.Errorf("Expected CacheMisses=1, got %d", snapshot.CacheMisses)
	}
	if snapshot.CacheSize != 100 {
		t.Errorf("Expected CacheSize=100, got %d", snapshot.CacheSize)
	}
}

func TestMetricsErrorTracking(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordNetworkError()
	metrics.RecordTimeoutError()
	metrics.RecordTimeoutError()
	metrics.RecordParseError()
	metrics.RecordRetryAttempt()

	snapshot := metrics.GetSnapshot()

	if snapshot.NetworkErrors != 1 {
		t.Errorf("Expected NetworkErrors=1, got %d", snapshot.NetworkErrors)
	}
	if snapshot.TimeoutErrors != 2 {
		t.Errorf("Expected TimeoutErrors=2, got %d", snapshot.TimeoutErrors)
	}
	if snapshot.ParseErrors != 1 {
		t.Errorf("Expected ParseErrors=1, got %d", snapshot.ParseErrors)
	}
	if snapshot.RetryAttempts != 1 {
		t.Errorf("Expected RetryAttempts=1, got %d", snapshot.RetryAttempts)
	}
}

func TestMetricsSummary(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordPageScraped(100 * time.Millisecond)
	metrics.RecordPageSkipped()
	metrics.RecordCacheHit()
	metrics.SetCacheSize(10)

	summary := metrics.Summary()

	// Check that summary contains expected sections
	expectedSections := []string{
		"Scraper Metrics Summary",
		"Uptime:",
		"Pages Scraped:",
		"Cache Performance:",
		"Error Statistics:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(summary, section) {
			t.Errorf("Expected summary to contain '%s'", section)
		}
	}
}

func TestMetricsPrometheusFormat(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordPageScraped(100 * time.Millisecond)
	metrics.RecordPageSkipped()
	metrics.RecordCacheHit()

	prometheus := metrics.PrometheusFormat()

	// Check Prometheus format structure
	expectedMetrics := []string{
		"scraper_urls_discovered",
		"scraper_pages_scraped",
		"scraper_pages_skipped",
		"scraper_cache_hits",
		"scraper_uptime_seconds",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(prometheus, metric) {
			t.Errorf("Expected Prometheus output to contain metric '%s'", metric)
		}
	}

	// Check for HELP and TYPE directives
	if !strings.Contains(prometheus, "# HELP") {
		t.Error("Expected Prometheus output to contain HELP directives")
	}
	if !strings.Contains(prometheus, "# TYPE") {
		t.Error("Expected Prometheus output to contain TYPE directives")
	}
}

func TestMetricsConcurrentAccess(t *testing.T) {
	metrics := NewMetrics()

	// Test concurrent writes
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				metrics.RecordPageScraped(10 * time.Millisecond)
				metrics.RecordCacheHit()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	snapshot := metrics.GetSnapshot()

	if snapshot.PagesScraped != 1000 {
		t.Errorf("Expected PagesScraped=1000, got %d (race condition?)", snapshot.PagesScraped)
	}
	if snapshot.CacheHits != 1000 {
		t.Errorf("Expected CacheHits=1000, got %d (race condition?)", snapshot.CacheHits)
	}
}

func TestMetricsHitRateCalculation(t *testing.T) {
	metrics := NewMetrics()

	// 3 hits, 1 miss = 75% hit rate
	metrics.RecordCacheHit()
	metrics.RecordCacheHit()
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()

	summary := metrics.Summary()

	if !strings.Contains(summary, "75.0%") {
		t.Errorf("Expected hit rate of 75%%, got summary: %s", summary)
	}
}

func TestMetricsAveragePageDuration(t *testing.T) {
	metrics := NewMetrics()

	// Record 3 scrapes: 100ms, 200ms, 300ms
	// Average should be 200ms
	metrics.RecordPageScraped(100 * time.Millisecond)
	metrics.RecordPageScraped(200 * time.Millisecond)
	metrics.RecordPageScraped(300 * time.Millisecond)

	snapshot := metrics.GetSnapshot()

	// Check average is approximately 200ms (within 10ms tolerance)
	expected := 200 * time.Millisecond
	diff := snapshot.AvgPageDuration - expected
	if diff < -10*time.Millisecond || diff > 10*time.Millisecond {
		t.Errorf("Expected AvgPageDuration~200ms, got %v", snapshot.AvgPageDuration)
	}
}

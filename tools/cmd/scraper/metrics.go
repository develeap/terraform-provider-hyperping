package main

import (
	"fmt"
	"sync"
	"time"
)

// Metrics tracks scraper performance and health
type Metrics struct {
	mu sync.RWMutex

	// Scraping metrics
	URLsDiscovered  int64
	PagesScraped    int64
	PagesSkipped    int64
	PagesFailed     int64
	TotalDuration   time.Duration
	AvgPageDuration time.Duration

	// Cache metrics
	CacheHits   int64
	CacheMisses int64
	CacheSize   int64

	// Error metrics
	NetworkErrors int64
	TimeoutErrors int64
	ParseErrors   int64
	RetryAttempts int64

	// Resource metrics
	BrowserRestarts int64
	MemoryUsageMB   int64

	// Timestamps
	StartTime     time.Time
	LastScrapedAt time.Time
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// RecordURLDiscovered increments URLs discovered counter
func (m *Metrics) RecordURLDiscovered() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.URLsDiscovered++
}

// RecordPageScraped increments successful scrape counter
func (m *Metrics) RecordPageScraped(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PagesScraped++
	m.LastScrapedAt = time.Now()

	// Update average duration
	if m.PagesScraped == 1 {
		m.AvgPageDuration = duration
	} else {
		// Running average
		m.AvgPageDuration = time.Duration(
			(int64(m.AvgPageDuration)*(m.PagesScraped-1) + int64(duration)) / m.PagesScraped,
		)
	}
}

// RecordPageSkipped increments skipped counter
func (m *Metrics) RecordPageSkipped() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PagesSkipped++
}

// RecordPageFailed increments failed counter
func (m *Metrics) RecordPageFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PagesFailed++
}

// RecordCacheHit increments cache hit counter
func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheHits++
}

// RecordCacheMiss increments cache miss counter
func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheMisses++
}

// SetCacheSize updates cache size metric
func (m *Metrics) SetCacheSize(size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CacheSize = size
}

// RecordNetworkError increments network error counter
func (m *Metrics) RecordNetworkError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.NetworkErrors++
}

// RecordTimeoutError increments timeout error counter
func (m *Metrics) RecordTimeoutError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TimeoutErrors++
}

// RecordParseError increments parse error counter
func (m *Metrics) RecordParseError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ParseErrors++
}

// RecordRetryAttempt increments retry attempt counter
func (m *Metrics) RecordRetryAttempt() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RetryAttempts++
}

// RecordBrowserRestart increments browser restart counter
func (m *Metrics) RecordBrowserRestart() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BrowserRestarts++
}

// SetMemoryUsage updates memory usage metric
func (m *Metrics) SetMemoryUsage(mb int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MemoryUsageMB = mb
}

// SetTotalDuration updates total scraping duration
func (m *Metrics) SetTotalDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalDuration = duration
}

// GetSnapshot returns a copy of current metrics (thread-safe)
func (m *Metrics) GetSnapshot() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m
}

// Summary returns a human-readable summary
func (m *Metrics) Summary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.StartTime).Round(time.Second)

	summary := fmt.Sprintf(`
Scraper Metrics Summary
=======================
Uptime:              %v
URLs Discovered:     %d
Pages Scraped:       %d
Pages Skipped:       %d
Pages Failed:        %d
Total Duration:      %v
Avg Page Duration:   %v

Cache Performance:
  Hits:              %d
  Misses:            %d
  Size:              %d entries
  Hit Rate:          %.1f%%

Error Statistics:
  Network Errors:    %d
  Timeout Errors:    %d
  Parse Errors:      %d
  Retry Attempts:    %d

Resources:
  Browser Restarts:  %d
  Memory Usage:      %d MB
`,
		uptime,
		m.URLsDiscovered,
		m.PagesScraped,
		m.PagesSkipped,
		m.PagesFailed,
		m.TotalDuration.Round(time.Second),
		m.AvgPageDuration.Round(time.Millisecond),
		m.CacheHits,
		m.CacheMisses,
		m.CacheSize,
		m.calculateHitRate(),
		m.NetworkErrors,
		m.TimeoutErrors,
		m.ParseErrors,
		m.RetryAttempts,
		m.BrowserRestarts,
		m.MemoryUsageMB,
	)

	return summary
}

// calculateHitRate calculates cache hit rate percentage
func (m *Metrics) calculateHitRate() float64 {
	total := m.CacheHits + m.CacheMisses
	if total == 0 {
		return 0.0
	}
	return (float64(m.CacheHits) / float64(total)) * 100
}

// PrometheusFormat returns metrics in Prometheus exposition format
func (m *Metrics) PrometheusFormat() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf(`# HELP scraper_urls_discovered Total number of URLs discovered
# TYPE scraper_urls_discovered counter
scraper_urls_discovered %d

# HELP scraper_pages_scraped Total number of pages successfully scraped
# TYPE scraper_pages_scraped counter
scraper_pages_scraped %d

# HELP scraper_pages_skipped Total number of pages skipped (unchanged)
# TYPE scraper_pages_skipped counter
scraper_pages_skipped %d

# HELP scraper_pages_failed Total number of failed page scrapes
# TYPE scraper_pages_failed counter
scraper_pages_failed %d

# HELP scraper_avg_page_duration_seconds Average page scrape duration in seconds
# TYPE scraper_avg_page_duration_seconds gauge
scraper_avg_page_duration_seconds %.3f

# HELP scraper_cache_hits Total number of cache hits
# TYPE scraper_cache_hits counter
scraper_cache_hits %d

# HELP scraper_cache_misses Total number of cache misses
# TYPE scraper_cache_misses counter
scraper_cache_misses %d

# HELP scraper_cache_size Current cache size (number of entries)
# TYPE scraper_cache_size gauge
scraper_cache_size %d

# HELP scraper_network_errors Total number of network errors
# TYPE scraper_network_errors counter
scraper_network_errors %d

# HELP scraper_timeout_errors Total number of timeout errors
# TYPE scraper_timeout_errors counter
scraper_timeout_errors %d

# HELP scraper_parse_errors Total number of parse errors
# TYPE scraper_parse_errors counter
scraper_parse_errors %d

# HELP scraper_retry_attempts Total number of retry attempts
# TYPE scraper_retry_attempts counter
scraper_retry_attempts %d

# HELP scraper_browser_restarts Total number of browser restarts
# TYPE scraper_browser_restarts counter
scraper_browser_restarts %d

# HELP scraper_memory_usage_mb Current memory usage in MB
# TYPE scraper_memory_usage_mb gauge
scraper_memory_usage_mb %d

# HELP scraper_uptime_seconds Scraper uptime in seconds
# TYPE scraper_uptime_seconds gauge
scraper_uptime_seconds %.0f
`,
		m.URLsDiscovered,
		m.PagesScraped,
		m.PagesSkipped,
		m.PagesFailed,
		m.AvgPageDuration.Seconds(),
		m.CacheHits,
		m.CacheMisses,
		m.CacheSize,
		m.NetworkErrors,
		m.TimeoutErrors,
		m.ParseErrors,
		m.RetryAttempts,
		m.BrowserRestarts,
		m.MemoryUsageMB,
		time.Since(m.StartTime).Seconds(),
	)
}

// Global metrics instance
var globalMetrics *Metrics

// InitMetrics initializes the global metrics instance
func InitMetrics() {
	globalMetrics = NewMetrics()
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	if globalMetrics == nil {
		InitMetrics()
	}
	return globalMetrics
}

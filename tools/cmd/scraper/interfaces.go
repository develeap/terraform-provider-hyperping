package main

import (
	"context"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/go-rod/rod"
)

// BrowserManager abstracts browser lifecycle operations
type BrowserManager interface {
	Launch(config ScraperConfig) error
	GetPage() (*rod.Page, error)
	Close() error
}

// CacheManager abstracts cache operations
type CacheManager interface {
	Load(filename string) (Cache, error)
	Save(filename string, cache Cache) error
	Update(cache *Cache, filename string, pageData *extractor.PageData)
	HasChanged(cache Cache, filename string, newContent string) bool
	Compare(oldCache, newCache Cache) map[string]interface{}
	GetStats(cache Cache) map[string]interface{}
}

// URLDiscoverer abstracts URL discovery operations
type URLDiscoverer interface {
	Discover(ctx context.Context, browser *rod.Browser, baseURL string) ([]DiscoveredURL, error)
	FilterNew(discovered []DiscoveredURL, cache Cache) []DiscoveredURL
	ToFilename(url string) string
}

// PageScraper abstracts page scraping operations
type PageScraper interface {
	Scrape(ctx context.Context, page *rod.Page, url string, timeout time.Duration) (*extractor.PageData, error)
	ScrapeWithRetry(ctx context.Context, page *rod.Page, url string, maxRetries int, timeout time.Duration) (*extractor.PageData, error)
}

// SnapshotStorage abstracts snapshot storage operations
type SnapshotStorage interface {
	Save(timestamp time.Time, pages map[string]*extractor.PageData) error
	GetLatest() (string, error)
	Load(snapshotDir string) (map[string]*extractor.PageData, error)
	Compare(oldSnapshot, newSnapshot string) ([]APIDiff, error)
	Cleanup(keep int) error
}

// DiffReporter abstracts diff reporting operations
type DiffReporter interface {
	Generate(diffs []APIDiff, timestamp time.Time) DiffReport
	Save(report DiffReport, filename string) error
	Format(report DiffReport) string
}

// GitHubIntegration abstracts GitHub operations
type GitHubIntegration interface {
	CreateIssue(report DiffReport, snapshotURL string) error
	CreateLabelsIfNeeded() error
	Preview(report DiffReport, snapshotURL string)
}

// ConfigProvider abstracts configuration access
type ConfigProvider interface {
	Get() ScraperConfig
	Validate() error
}

// MetricsCollector abstracts metrics collection
type MetricsCollector interface {
	RecordPageScraped(duration time.Duration)
	RecordPageSkipped()
	RecordPageFailed()
	RecordCacheHit()
	RecordCacheMiss()
	RecordNetworkError()
	RecordTimeoutError()
	RecordRetryAttempt()
	GetSnapshot() Metrics
	Summary() string
	PrometheusFormat() string
}

// LoggerInterface abstracts logging operations
type LoggerInterface interface {
	Debug(message string, fields ...map[string]interface{})
	Info(message string, fields ...map[string]interface{})
	Warn(message string, fields ...map[string]interface{})
	Error(message string, fields ...map[string]interface{})
	SetLevel(level LogLevel)
}

// Scraper represents the main scraper with injected dependencies
type Scraper struct {
	config     ScraperConfig
	cache      CacheManager
	discoverer URLDiscoverer
	scraper    PageScraper
	snapshots  SnapshotStorage
	differ     DiffReporter
	github     GitHubIntegration
	metrics    MetricsCollector
	logger     LoggerInterface
}

// NewScraper creates a scraper with dependency injection
func NewScraper(
	config ScraperConfig,
	cache CacheManager,
	discoverer URLDiscoverer,
	scraper PageScraper,
	snapshots SnapshotStorage,
	differ DiffReporter,
	github GitHubIntegration,
	metrics MetricsCollector,
	logger LoggerInterface,
) *Scraper {
	return &Scraper{
		config:     config,
		cache:      cache,
		discoverer: discoverer,
		scraper:    scraper,
		snapshots:  snapshots,
		differ:     differ,
		github:     github,
		metrics:    metrics,
		logger:     logger,
	}
}

// Run executes the scraping workflow
func (s *Scraper) Run(ctx context.Context) error {
	s.logger.Info("Starting scraper")

	// Implementation would use injected dependencies
	// This provides a clean architecture for testing

	return nil
}

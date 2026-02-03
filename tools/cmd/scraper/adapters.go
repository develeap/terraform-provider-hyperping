package main

import (
	"context"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/go-rod/rod"
)

// DefaultCacheManager wraps existing cache functions
type DefaultCacheManager struct{}

func (d *DefaultCacheManager) Load(filename string) (Cache, error) {
	return LoadCache(filename)
}

func (d *DefaultCacheManager) Save(filename string, cache Cache) error {
	return SaveCache(filename, cache)
}

func (d *DefaultCacheManager) Update(cache *Cache, filename string, pageData *extractor.PageData) {
	UpdateCache(cache, filename, pageData)
}

func (d *DefaultCacheManager) HasChanged(cache Cache, filename string, newContent string) bool {
	return HasChanged(cache, filename, newContent)
}

func (d *DefaultCacheManager) Compare(oldCache, newCache Cache) map[string]interface{} {
	return CompareCaches(oldCache, newCache)
}

func (d *DefaultCacheManager) GetStats(cache Cache) map[string]interface{} {
	return GetCacheStats(cache)
}

// DefaultURLDiscoverer wraps existing discovery functions
type DefaultURLDiscoverer struct{}

func (d *DefaultURLDiscoverer) Discover(ctx context.Context, browser *rod.Browser, baseURL string) ([]DiscoveredURL, error) {
	return DiscoverURLs(ctx, browser, baseURL)
}

func (d *DefaultURLDiscoverer) FilterNew(discovered []DiscoveredURL, cache Cache) []DiscoveredURL {
	return FilterNewURLs(discovered, cache)
}

func (d *DefaultURLDiscoverer) ToFilename(url string) string {
	return URLToFilename(url)
}

// DefaultPageScraper wraps existing scraping functions
type DefaultPageScraper struct{}

func (d *DefaultPageScraper) Scrape(ctx context.Context, page *rod.Page, url string, timeout time.Duration) (*extractor.PageData, error) {
	return scrapePage(ctx, page, url, timeout)
}

func (d *DefaultPageScraper) ScrapeWithRetry(ctx context.Context, page *rod.Page, url string, maxRetries int, timeout time.Duration) (*extractor.PageData, error) {
	return scrapeWithRetry(ctx, page, url, maxRetries, timeout)
}

// DefaultSnapshotStorage wraps SnapshotManager
type DefaultSnapshotStorage struct {
	manager *SnapshotManager
}

func NewDefaultSnapshotStorage(baseDir string) *DefaultSnapshotStorage {
	return &DefaultSnapshotStorage{
		manager: NewSnapshotManager(baseDir),
	}
}

func (d *DefaultSnapshotStorage) Save(timestamp time.Time, pages map[string]*extractor.PageData) error {
	return d.manager.SaveSnapshot(timestamp, pages)
}

func (d *DefaultSnapshotStorage) GetLatest() (string, error) {
	return d.manager.GetLatestSnapshot()
}

func (d *DefaultSnapshotStorage) Load(snapshotDir string) (map[string]*extractor.PageData, error) {
	return d.manager.LoadSnapshot(snapshotDir)
}

func (d *DefaultSnapshotStorage) Compare(oldSnapshot, newSnapshot string) ([]APIDiff, error) {
	return d.manager.CompareSnapshots(oldSnapshot, newSnapshot)
}

func (d *DefaultSnapshotStorage) Cleanup(keep int) error {
	return d.manager.CleanupOldSnapshots(keep)
}

// DefaultDiffReporter wraps diff functions
type DefaultDiffReporter struct{}

func (d *DefaultDiffReporter) Generate(diffs []APIDiff, timestamp time.Time) DiffReport {
	return GenerateDiffReport(diffs, timestamp)
}

func (d *DefaultDiffReporter) Save(report DiffReport, filename string) error {
	return SaveDiffReport(report, filename)
}

func (d *DefaultDiffReporter) Format(report DiffReport) string {
	return FormatDiffAsMarkdown(report)
}

// DefaultGitHubIntegration wraps GitHub client
type DefaultGitHubIntegration struct {
	client *GitHubClient
}

func NewDefaultGitHubIntegration() (*DefaultGitHubIntegration, error) {
	client, err := LoadGitHubConfig()
	if err != nil {
		return nil, err
	}
	return &DefaultGitHubIntegration{client: client}, nil
}

func (d *DefaultGitHubIntegration) CreateIssue(report DiffReport, snapshotURL string) error {
	return d.client.CreateIssue(report, snapshotURL)
}

func (d *DefaultGitHubIntegration) CreateLabelsIfNeeded() error {
	return d.client.CreateLabelsIfNeeded()
}

func (d *DefaultGitHubIntegration) Preview(report DiffReport, snapshotURL string) {
	PreviewGitHubIssue(report, snapshotURL)
}

// DefaultConfigProvider wraps config
type DefaultConfigProvider struct {
	config ScraperConfig
}

func NewDefaultConfigProvider() *DefaultConfigProvider {
	return &DefaultConfigProvider{
		config: DefaultConfig(),
	}
}

func (d *DefaultConfigProvider) Get() ScraperConfig {
	return d.config
}

func (d *DefaultConfigProvider) Validate() error {
	// Add validation logic if needed
	return nil
}

// LoggerAdapter wraps Logger to satisfy LoggerInterface
type LoggerAdapter struct {
	logger *Logger
}

func NewLoggerAdapter(logger *Logger) *LoggerAdapter {
	return &LoggerAdapter{logger: logger}
}

func (l *LoggerAdapter) Debug(message string, fields ...map[string]interface{}) {
	l.logger.Debug(message, fields...)
}

func (l *LoggerAdapter) Info(message string, fields ...map[string]interface{}) {
	l.logger.Info(message, fields...)
}

func (l *LoggerAdapter) Warn(message string, fields ...map[string]interface{}) {
	l.logger.Warn(message, fields...)
}

func (l *LoggerAdapter) Error(message string, fields ...map[string]interface{}) {
	l.logger.Error(message, fields...)
}

func (l *LoggerAdapter) SetLevel(level LogLevel) {
	l.logger.SetLevel(level)
}

// MetricsAdapter wraps Metrics to satisfy MetricsCollector
type MetricsAdapter struct {
	metrics *Metrics
}

func NewMetricsAdapter(metrics *Metrics) *MetricsAdapter {
	return &MetricsAdapter{metrics: metrics}
}

func (m *MetricsAdapter) RecordPageScraped(duration time.Duration) {
	m.metrics.RecordPageScraped(duration)
}

func (m *MetricsAdapter) RecordPageSkipped() {
	m.metrics.RecordPageSkipped()
}

func (m *MetricsAdapter) RecordPageFailed() {
	m.metrics.RecordPageFailed()
}

func (m *MetricsAdapter) RecordCacheHit() {
	m.metrics.RecordCacheHit()
}

func (m *MetricsAdapter) RecordCacheMiss() {
	m.metrics.RecordCacheMiss()
}

func (m *MetricsAdapter) RecordNetworkError() {
	m.metrics.RecordNetworkError()
}

func (m *MetricsAdapter) RecordTimeoutError() {
	m.metrics.RecordTimeoutError()
}

func (m *MetricsAdapter) RecordRetryAttempt() {
	m.metrics.RecordRetryAttempt()
}

func (m *MetricsAdapter) GetSnapshot() Metrics {
	return m.metrics.GetSnapshot()
}

func (m *MetricsAdapter) Summary() string {
	return m.metrics.Summary()
}

func (m *MetricsAdapter) PrometheusFormat() string {
	return m.metrics.PrometheusFormat()
}

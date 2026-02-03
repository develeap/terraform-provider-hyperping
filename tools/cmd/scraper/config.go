package main

import "time"

// Configuration constants
const (
	// Scraping configuration
	DefaultBaseURL          = "https://hyperping.com/docs/api"
	DefaultOutputDir        = "docs_scraped"
	DefaultCacheFile        = ".scraper_cache.json"
	DefaultSnapshotsDir     = "snapshots"
	DefaultRateLimit        = 1.0 // requests per second
	DefaultTimeout          = 30 * time.Second
	DefaultRetries          = 3
	DefaultHeadless         = true
	DefaultResourceBlocking = true

	// Snapshot retention
	DefaultSnapshotRetention = 10 // Keep last 10 snapshots

	// API documentation structure
	HyperpingDocsBaseURL = "https://hyperping.com"
	NotionAPIDocsURL     = "https://hyperping.notion.site/Hyperping-API-1720a8d31cf380d4a7f4f58e36fa10dd"
)

// ScraperConfig holds configuration for the scraper
type ScraperConfig struct {
	BaseURL          string        `json:"base_url"`
	OutputDir        string        `json:"output_dir"`
	CacheFile        string        `json:"cache_file"`
	SnapshotsDir     string        `json:"snapshots_dir"`
	RateLimit        float64       `json:"rate_limit"`        // Requests per second
	Timeout          time.Duration `json:"timeout"`           // Per-page timeout
	Retries          int           `json:"retries"`           // Max retry attempts
	Headless         bool          `json:"headless"`          // Run browser headless
	ResourceBlocking bool          `json:"resource_blocking"` // Block images/CSS/fonts
}

// DefaultConfig returns sensible defaults for the scraper
func DefaultConfig() ScraperConfig {
	return ScraperConfig{
		BaseURL:          DefaultBaseURL,
		OutputDir:        DefaultOutputDir,
		CacheFile:        DefaultCacheFile,
		SnapshotsDir:     DefaultSnapshotsDir,
		RateLimit:        DefaultRateLimit,
		Timeout:          DefaultTimeout,
		Retries:          DefaultRetries,
		Headless:         DefaultHeadless,
		ResourceBlocking: DefaultResourceBlocking,
	}
}

// GitHubConfig holds GitHub integration configuration
type GitHubConfig struct {
	Token string
	Owner string // e.g., "develeap"
	Repo  string // e.g., "terraform-provider-hyperping"
}

// RequiredLabels defines all labels that should exist in the repository
var RequiredLabels = map[string]string{
	"api-change":       "0E8A16", // green
	"breaking-change":  "D93F0B", // red
	"non-breaking":     "0E8A16", // green
	"automated":        "FBCA04", // yellow
	"monitors-api":     "1D76DB", // blue
	"statuspages-api":  "1D76DB", // blue
	"incidents-api":    "1D76DB", // blue
	"outages-api":      "1D76DB", // blue
	"maintenance-api":  "1D76DB", // blue
	"healthchecks-api": "1D76DB", // blue
	"reports-api":      "1D76DB", // blue
}

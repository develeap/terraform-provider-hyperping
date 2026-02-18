package main

import "time"

// CacheEntry stores metadata for change detection.
type CacheEntry struct {
	Filename     string    `json:"filename"`
	URL          string    `json:"url"`
	ContentHash  string    `json:"content_hash"`
	Size         int       `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// Cache stores all cache entries with creation timestamp.
type Cache struct {
	Entries   map[string]CacheEntry `json:"entries"`
	CreatedAt time.Time             `json:"created_at"`
}

// ScrapeStats holds statistics from a single scraping run.
type ScrapeStats struct {
	Scraped  int
	Skipped  int
	Failed   int
	Duration time.Duration
}

// DiscoveredURL is a local type used by cache helpers and tests.
// New code should prefer discovery.DiscoveredURL.
type DiscoveredURL struct {
	URL     string
	Section string
	Method  string
}

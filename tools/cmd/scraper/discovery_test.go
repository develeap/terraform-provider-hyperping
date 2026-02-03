package main

import (
	"testing"
	"time"
)

func TestURLToFilename(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "monitors create",
			url:      "https://hyperping.com/docs/api/monitors/create",
			expected: "monitors_create.json",
		},
		{
			name:     "monitors list",
			url:      "https://hyperping.com/docs/api/monitors",
			expected: "monitors.json",
		},
		{
			name:     "incidents section",
			url:      "https://hyperping.com/docs/api/incidents",
			expected: "incidents.json",
		},
		{
			name:     "root path",
			url:      "https://hyperping.com/docs/api/",
			expected: "index.json",
		},
		{
			name:     "notion url",
			url:      "https://hyperping.notion.site/API-123",
			expected: "API-123.json",
		},
		{
			name:     "url with query params",
			url:      "https://hyperping.com/docs/api/monitors?page=1&limit=10",
			expected: "monitors_page_1_limit_10.json",
		},
		{
			name:     "url with special chars",
			url:      "https://hyperping.com/docs/api/monitors&create=true",
			expected: "monitors_create_true.json",
		},
		{
			name:     "nested path",
			url:      "https://hyperping.com/docs/api/statuspages/subscribers/email",
			expected: "statuspages_subscribers_email.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := URLToFilename(tt.url)
			if result != tt.expected {
				t.Errorf("URLToFilename(%s) = %s, expected %s", tt.url, result, tt.expected)
			}
		})
	}
}

func TestFilterNewURLs(t *testing.T) {
	cache := Cache{
		Entries: map[string]CacheEntry{
			"monitors.json": {
				Filename:     "monitors.json",
				ContentHash:  "hash1",
				LastModified: time.Now(),
			},
			"incidents.json": {
				Filename:     "incidents.json",
				ContentHash:  "hash2",
				LastModified: time.Now(),
			},
		},
		CreatedAt: time.Now(),
	}

	discovered := []DiscoveredURL{
		{URL: "https://hyperping.com/docs/api/monitors", Section: "monitors", Method: "GET"},
		{URL: "https://hyperping.com/docs/api/incidents", Section: "incidents", Method: "GET"},
		{URL: "https://hyperping.com/docs/api/statuspages", Section: "statuspages", Method: "GET"},
		{URL: "https://hyperping.com/docs/api/outages", Section: "outages", Method: "GET"},
	}

	newURLs := FilterNewURLs(discovered, cache)

	// Should only return statuspages and outages (not in cache)
	if len(newURLs) != 2 {
		t.Errorf("Expected 2 new URLs, got %d", len(newURLs))
	}

	// Verify correct URLs are returned
	found := make(map[string]bool)
	for _, url := range newURLs {
		found[url.Section] = true
	}

	if !found["statuspages"] {
		t.Error("Expected statuspages in new URLs")
	}
	if !found["outages"] {
		t.Error("Expected outages in new URLs")
	}
	if found["monitors"] {
		t.Error("Did not expect monitors in new URLs (already in cache)")
	}
	if found["incidents"] {
		t.Error("Did not expect incidents in new URLs (already in cache)")
	}
}

func TestFilterNewURLs_EmptyCache(t *testing.T) {
	cache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	discovered := []DiscoveredURL{
		{URL: "https://hyperping.com/docs/api/monitors", Section: "monitors", Method: "GET"},
		{URL: "https://hyperping.com/docs/api/incidents", Section: "incidents", Method: "GET"},
	}

	newURLs := FilterNewURLs(discovered, cache)

	// All URLs should be new
	if len(newURLs) != 2 {
		t.Errorf("Expected 2 new URLs, got %d", len(newURLs))
	}
}

func TestFilterNewURLs_AllCached(t *testing.T) {
	cache := Cache{
		Entries: map[string]CacheEntry{
			"monitors.json":  {Filename: "monitors.json", ContentHash: "hash1", LastModified: time.Now()},
			"incidents.json": {Filename: "incidents.json", ContentHash: "hash2", LastModified: time.Now()},
		},
		CreatedAt: time.Now(),
	}

	discovered := []DiscoveredURL{
		{URL: "https://hyperping.com/docs/api/monitors", Section: "monitors", Method: "GET"},
		{URL: "https://hyperping.com/docs/api/incidents", Section: "incidents", Method: "GET"},
	}

	newURLs := FilterNewURLs(discovered, cache)

	// No URLs should be new
	if len(newURLs) != 0 {
		t.Errorf("Expected 0 new URLs, got %d", len(newURLs))
	}
}

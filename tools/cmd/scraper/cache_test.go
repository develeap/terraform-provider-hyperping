package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
)

func TestUpdateCache(t *testing.T) {
	cache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	pageData := &extractor.PageData{
		URL:       "https://example.com/api",
		Title:     "API Docs",
		Text:      "Test content for hashing",
		HTML:      "<html>test</html>",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	UpdateCache(&cache, "api.json", pageData)

	if len(cache.Entries) != 1 {
		t.Errorf("Expected 1 cache entry, got %d", len(cache.Entries))
	}

	entry, exists := cache.Entries["api.json"]
	if !exists {
		t.Fatal("Expected api.json in cache")
	}

	if entry.Filename != "api.json" {
		t.Errorf("Expected filename=api.json, got %s", entry.Filename)
	}
	if entry.URL != pageData.URL {
		t.Errorf("Expected URL=%s, got %s", pageData.URL, entry.URL)
	}
	if entry.Size != len(pageData.Text) {
		t.Errorf("Expected size=%d, got %d", len(pageData.Text), entry.Size)
	}
	if entry.ContentHash == "" {
		t.Error("Expected non-empty content hash")
	}
}

func TestHasChanged_NewFile(t *testing.T) {
	cache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	changed := HasChanged(cache, "new_file.json", "new content")
	if !changed {
		t.Error("Expected HasChanged=true for new file")
	}
}

func TestHasChanged_SameContent(t *testing.T) {
	content := "unchanged content"

	// Create cache with entry
	cache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	pageData := &extractor.PageData{
		URL:       "https://example.com",
		Text:      content,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	UpdateCache(&cache, "file.json", pageData)

	// Check same content
	changed := HasChanged(cache, "file.json", content)
	if changed {
		t.Error("Expected HasChanged=false for same content")
	}
}

func TestHasChanged_DifferentContent(t *testing.T) {
	// Create cache with entry
	cache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	pageData := &extractor.PageData{
		URL:       "https://example.com",
		Text:      "old content",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	UpdateCache(&cache, "file.json", pageData)

	// Check different content
	changed := HasChanged(cache, "file.json", "new content")
	if !changed {
		t.Error("Expected HasChanged=true for different content")
	}
}

func TestGetCacheStats(t *testing.T) {
	now := time.Now()
	cache := Cache{
		Entries: map[string]CacheEntry{
			"file1.json": {
				Filename:     "file1.json",
				URL:          "https://example.com/1",
				ContentHash:  "hash1",
				Size:         100,
				LastModified: now.Add(-1 * time.Hour),
			},
			"file2.json": {
				Filename:     "file2.json",
				URL:          "https://example.com/2",
				ContentHash:  "hash2",
				Size:         200,
				LastModified: now,
			},
		},
		CreatedAt: now.Add(-2 * time.Hour),
	}

	stats := GetCacheStats(cache)

	if stats["total_entries"] != 2 {
		t.Errorf("Expected total_entries=2, got %v", stats["total_entries"])
	}
	if stats["total_size"] != 300 {
		t.Errorf("Expected total_size=300, got %v", stats["total_size"])
	}

	// Check that oldest/newest are set
	if stats["oldest_entry"] == nil {
		t.Error("Expected oldest_entry to be set")
	}
	if stats["newest_entry"] == nil {
		t.Error("Expected newest_entry to be set")
	}
	if stats["cache_age"] == nil {
		t.Error("Expected cache_age to be set")
	}
}

func TestCompareCaches(t *testing.T) {
	oldCache := Cache{
		Entries: map[string]CacheEntry{
			"unchanged.json": {Filename: "unchanged.json", ContentHash: "hash1", LastModified: time.Now()},
			"modified.json":  {Filename: "modified.json", ContentHash: "oldhash", LastModified: time.Now()},
			"deleted.json":   {Filename: "deleted.json", ContentHash: "hash3", LastModified: time.Now()},
		},
		CreatedAt: time.Now(),
	}

	newCache := Cache{
		Entries: map[string]CacheEntry{
			"unchanged.json": {Filename: "unchanged.json", ContentHash: "hash1", LastModified: time.Now()}, // Same hash
			"modified.json":  {Filename: "modified.json", ContentHash: "newhash", LastModified: time.Now()},  // Different hash
			"added.json":     {Filename: "added.json", ContentHash: "hash4", LastModified: time.Now()},       // New file
		},
		CreatedAt: time.Now(),
	}

	stats := CompareCaches(oldCache, newCache)

	if stats["unchanged"] != 1 {
		t.Errorf("Expected unchanged=1, got %v", stats["unchanged"])
	}
	if stats["modified"] != 1 {
		t.Errorf("Expected modified=1, got %v", stats["modified"])
	}
	if stats["added"] != 1 {
		t.Errorf("Expected added=1, got %v", stats["added"])
	}
	if stats["deleted"] != 1 {
		t.Errorf("Expected deleted=1, got %v", stats["deleted"])
	}
	if stats["total"] != 3 {
		t.Errorf("Expected total=3, got %v", stats["total"])
	}
}

func TestSaveCache(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "test_cache.json")

	cache := Cache{
		Entries: map[string]CacheEntry{
			"test.json": {
				Filename:     "test.json",
				URL:          "https://example.com",
				ContentHash:  "hash123",
				Size:         100,
				LastModified: time.Now(),
			},
		},
		CreatedAt: time.Now(),
	}

	err := SaveCache(cacheFile, cache)
	if err != nil {
		t.Fatalf("SaveCache failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Errorf("Expected cache file to exist at %s", cacheFile)
	}

	// Load and verify
	loaded, err := LoadCache(cacheFile)
	if err != nil {
		t.Fatalf("LoadCache failed: %v", err)
	}

	if len(loaded.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(loaded.Entries))
	}

	entry, exists := loaded.Entries["test.json"]
	if !exists {
		t.Fatal("Expected test.json in loaded cache")
	}

	if entry.ContentHash != "hash123" {
		t.Errorf("Expected hash=hash123, got %s", entry.ContentHash)
	}
}

func TestLoadCache_NoFile(t *testing.T) {
	cache, err := LoadCache("/nonexistent/cache.json")
	if err != nil {
		t.Fatalf("LoadCache should not error on missing file: %v", err)
	}

	if len(cache.Entries) != 0 {
		t.Errorf("Expected empty cache for missing file, got %d entries", len(cache.Entries))
	}
}

func TestBuildCacheFromDisk(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	page1 := &extractor.PageData{
		URL:       "https://example.com/1",
		Title:     "Page 1",
		Text:      "Content 1",
		HTML:      "<html>1</html>",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	page2 := &extractor.PageData{
		URL:       "https://example.com/2",
		Title:     "Page 2",
		Text:      "Content 2",
		HTML:      "<html>2</html>",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	savePageData(filepath.Join(tmpDir, "page1.json"), page1)
	savePageData(filepath.Join(tmpDir, "page2.json"), page2)

	cache, err := BuildCacheFromDisk(tmpDir)
	if err != nil {
		t.Fatalf("BuildCacheFromDisk failed: %v", err)
	}

	if len(cache.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(cache.Entries))
	}

	if _, exists := cache.Entries["page1.json"]; !exists {
		t.Error("Expected page1.json in cache")
	}
	if _, exists := cache.Entries["page2.json"]; !exists {
		t.Error("Expected page2.json in cache")
	}
}

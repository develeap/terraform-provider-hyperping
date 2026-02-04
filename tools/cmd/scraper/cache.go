package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
)

// LoadCache reads the cache file from disk
func LoadCache(filename string) (Cache, error) {
	data, err := os.ReadFile(filename) // #nosec G304 -- filename is from internal config
	if err != nil {
		if os.IsNotExist(err) {
			// Cache doesn't exist yet, return empty
			log.Println("📦 No existing cache found, creating new cache")
			return Cache{
				Entries:   make(map[string]CacheEntry),
				CreatedAt: time.Now(),
			}, nil
		}
		return Cache{}, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return Cache{}, fmt.Errorf("failed to parse cache file: %w", err)
	}

	log.Printf("📦 Loaded cache with %d entries from %s\n", len(cache.Entries), cache.CreatedAt.Format("2006-01-02 15:04:05"))
	return cache, nil
}

// SaveCache writes the cache to disk using atomic write (crash-safe)
func SaveCache(filename string, cache Cache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := utils.AtomicWriteFile(filename, data, utils.FilePermPrivate); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	log.Printf("💾 Cache saved with %d entries\n", len(cache.Entries))
	return nil
}

// UpdateCache adds or updates a cache entry for a scraped page
func UpdateCache(cache *Cache, filename string, pageData *extractor.PageData) {
	contentHash := utils.HashContent(pageData.Text)

	// Parse timestamp string to time.Time
	timestamp, err := time.Parse(time.RFC3339, pageData.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	entry := CacheEntry{
		Filename:     filename,
		URL:          pageData.URL,
		ContentHash:  contentHash,
		Size:         len(pageData.Text),
		LastModified: timestamp,
	}

	cache.Entries[filename] = entry
}

// HasChanged checks if a page's content has changed compared to cache
func HasChanged(cache Cache, filename string, newContent string) bool {
	entry, exists := cache.Entries[filename]
	if !exists {
		// Not in cache, consider it changed (new)
		return true
	}

	newHash := utils.HashContent(newContent)
	return entry.ContentHash != newHash
}

// BuildCacheFromDisk creates a cache from existing scraped files
// Useful for recovering from a crashed scraper or manual inspection
func BuildCacheFromDisk(dir string) (Cache, error) {
	cache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return cache, fmt.Errorf("failed to list files: %w", err)
	}

	log.Printf("🔨 Building cache from %d files in %s\n", len(files), dir)

	for _, filepath := range files {
		filename := filepath[len(dir)+1:] // Remove directory prefix

		data, err := os.ReadFile(filepath) // #nosec G304 -- filepath is from internal glob
		if err != nil {
			log.Printf("   ⚠️  Failed to read %s: %v\n", filename, err)
			continue
		}

		var pageData extractor.PageData
		if err := json.Unmarshal(data, &pageData); err != nil {
			log.Printf("   ⚠️  Failed to parse %s: %v\n", filename, err)
			continue
		}

		contentHash := utils.HashContent(pageData.Text)

		// Parse timestamp string to time.Time
		timestamp, err := time.Parse(time.RFC3339, pageData.Timestamp)
		if err != nil {
			timestamp = time.Now()
		}

		cache.Entries[filename] = CacheEntry{
			Filename:     filename,
			URL:          pageData.URL,
			ContentHash:  contentHash,
			Size:         len(pageData.Text),
			LastModified: timestamp,
		}
	}

	log.Printf("✅ Cache built with %d entries\n", len(cache.Entries))
	return cache, nil
}

// GetCacheStats returns statistics about the cache
func GetCacheStats(cache Cache) map[string]interface{} {
	totalSize := 0
	oldestEntry := time.Now()
	newestEntry := time.Time{}

	for _, entry := range cache.Entries {
		totalSize += entry.Size
		if entry.LastModified.Before(oldestEntry) {
			oldestEntry = entry.LastModified
		}
		if entry.LastModified.After(newestEntry) {
			newestEntry = entry.LastModified
		}
	}

	return map[string]interface{}{
		"total_entries": len(cache.Entries),
		"total_size":    totalSize,
		"oldest_entry":  oldestEntry,
		"newest_entry":  newestEntry,
		"cache_age":     time.Since(cache.CreatedAt).Round(time.Second),
	}
}

// CompareCaches compares two caches and returns statistics
func CompareCaches(oldCache, newCache Cache) map[string]interface{} {
	unchanged := 0
	modified := 0
	added := 0
	deleted := 0

	// Check for modifications and additions
	for filename, newEntry := range newCache.Entries {
		oldEntry, exists := oldCache.Entries[filename]

		if !exists {
			added++
			continue
		}

		if oldEntry.ContentHash != newEntry.ContentHash {
			modified++
		} else {
			unchanged++
		}
	}

	// Check for deletions
	for filename := range oldCache.Entries {
		if _, exists := newCache.Entries[filename]; !exists {
			deleted++
		}
	}

	return map[string]interface{}{
		"unchanged": unchanged,
		"modified":  modified,
		"added":     added,
		"deleted":   deleted,
		"total":     len(newCache.Entries),
	}
}

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PageData struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	HTML      string    `json:"html"`
	Timestamp time.Time `json:"timestamp"`
}

type CacheEntry struct {
	Filename     string    `json:"filename"`
	URL          string    `json:"url"`
	ContentHash  string    `json:"content_hash"`
	Size         int       `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

type Cache struct {
	Entries   map[string]CacheEntry `json:"entries"`
	CreatedAt time.Time             `json:"created_at"`
}

func main() {
	fmt.Println("🔄 Hyperping API - Change Detection Test")
	fmt.Println(strings.Repeat("=", 60))

	cacheFile := ".scraper_cache.json"
	dir := "docs_scraped_mvp"

	// Load or create cache
	cache := loadCache(cacheFile)
	if len(cache.Entries) == 0 {
		fmt.Println("\n📝 First run - building cache...")
		cache = buildCache(dir)
		saveCache(cacheFile, cache)
		fmt.Printf("✅ Cache created with %d entries\n", len(cache.Entries))
		fmt.Println("\n💡 Run this tool again after re-scraping to detect changes")
		return
	}

	fmt.Printf("\n📊 Comparing with cache from: %s\n", cache.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("-", 60))

	// Build new cache from current files
	newCache := buildCache(dir)

	// Compare
	unchanged := 0
	modified := 0
	added := 0
	deleted := 0

	fmt.Println("\n📄 Change Detection Results:\n")

	// Check for modifications and additions
	for filename, newEntry := range newCache.Entries {
		oldEntry, exists := cache.Entries[filename]

		if !exists {
			fmt.Printf("  🆕 ADDED: %s (%d chars)\n", filename, newEntry.Size)
			added++
			continue
		}

		if oldEntry.ContentHash != newEntry.ContentHash {
			fmt.Printf("  📝 MODIFIED: %s\n", filename)
			fmt.Printf("      Old: %d chars (hash: %s...)\n", oldEntry.Size, oldEntry.ContentHash[:12])
			fmt.Printf("      New: %d chars (hash: %s...)\n", newEntry.Size, newEntry.ContentHash[:12])
			fmt.Printf("      Size change: %+d chars\n", newEntry.Size-oldEntry.Size)
			modified++
		} else {
			unchanged++
		}
	}

	// Check for deletions
	for filename := range cache.Entries {
		if _, exists := newCache.Entries[filename]; !exists {
			fmt.Printf("  ❌ DELETED: %s\n", filename)
			deleted++
		}
	}

	// Summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  ✅ Unchanged: %d files\n", unchanged)
	fmt.Printf("  📝 Modified:  %d files\n", modified)
	fmt.Printf("  🆕 Added:     %d files\n", added)
	fmt.Printf("  ❌ Deleted:   %d files\n", deleted)
	fmt.Printf("  📦 Total:     %d files\n", len(newCache.Entries))

	if modified == 0 && added == 0 && deleted == 0 {
		fmt.Println("\n✅ No changes detected - incremental scraping would skip all pages!")
		fmt.Printf("   Potential time saved: ~%ds (only 2-3s overhead for checking)\n",
			len(newCache.Entries)) // 1 sec per page
	} else {
		fmt.Printf("\n💡 Incremental scraping would re-scrape only %d pages\n", modified+added)
		fmt.Printf("   Time saved: ~%d%% (%d pages skipped)\n",
			(unchanged*100)/len(newCache.Entries),
			unchanged)
	}

	// Update cache
	fmt.Println("\n💾 Updating cache...")
	saveCache(cacheFile, newCache)
	fmt.Println("✅ Cache updated")
}

func buildCache(dir string) Cache {
	cache := Cache{
		Entries:   make(map[string]CacheEntry),
		CreatedAt: time.Now(),
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return cache
	}

	for _, filepath := range files {
		filename := filepath[len(dir)+1:] // Remove directory prefix

		data, err := os.ReadFile(filepath)
		if err != nil {
			continue
		}

		var pageData PageData
		if err := json.Unmarshal(data, &pageData); err != nil {
			continue
		}

		contentHash := hashContent(pageData.Text)

		cache.Entries[filename] = CacheEntry{
			Filename:     filename,
			URL:          pageData.URL,
			ContentHash:  contentHash,
			Size:         len(pageData.Text),
			LastModified: pageData.Timestamp,
		}
	}

	return cache
}

func loadCache(filename string) Cache {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Cache{Entries: make(map[string]CacheEntry)}
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return Cache{Entries: make(map[string]CacheEntry)}
	}

	return cache
}

func saveCache(filename string, cache Cache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

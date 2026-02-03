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

// SnapshotManager handles storing and retrieving API snapshots
type SnapshotManager struct {
	BaseDir string // snapshots/
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(baseDir string) *SnapshotManager {
	return &SnapshotManager{
		BaseDir: baseDir,
	}
}

// SaveSnapshot saves the current scraped data as a timestamped snapshot
func (sm *SnapshotManager) SaveSnapshot(timestamp time.Time, pages map[string]*extractor.PageData) error {
	// Create snapshot directory: snapshots/2026-02-03_10-30-00/
	snapshotDir := filepath.Join(sm.BaseDir, timestamp.Format("2006-01-02_15-04-05"))

	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Save each page
	for filename, pageData := range pages {
		filePath := filepath.Join(snapshotDir, filename)

		data, err := json.MarshalIndent(pageData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal %s: %w", filename, err)
		}

		if err := utils.AtomicWriteFile(filePath, data, utils.FilePermPrivate); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	log.Printf("📸 Snapshot saved: %s (%d files)\n", snapshotDir, len(pages))
	return nil
}

// GetLatestSnapshot returns the most recent snapshot directory
func (sm *SnapshotManager) GetLatestSnapshot() (string, error) {
	// List all snapshot directories
	entries, err := os.ReadDir(sm.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no snapshots exist yet")
		}
		return "", fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	// Filter directories only
	var snapshots []string
	for _, entry := range entries {
		if entry.IsDir() {
			snapshots = append(snapshots, entry.Name())
		}
	}

	if len(snapshots) == 0 {
		return "", fmt.Errorf("no snapshots found")
	}

	// Return the last one (they're sorted alphabetically by timestamp)
	latest := snapshots[len(snapshots)-1]
	return filepath.Join(sm.BaseDir, latest), nil
}

// LoadSnapshot loads all pages from a snapshot directory
func (sm *SnapshotManager) LoadSnapshot(snapshotDir string) (map[string]*extractor.PageData, error) {
	pages := make(map[string]*extractor.PageData)

	files, err := filepath.Glob(filepath.Join(snapshotDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	for _, filePath := range files {
		filename := filepath.Base(filePath)

		pageData, err := loadPageData(filePath)
		if err != nil {
			log.Printf("   ⚠️  Failed to load %s: %v\n", filename, err)
			continue
		}

		pages[filename] = pageData
	}

	return pages, nil
}

// CompareSnapshots compares two snapshots and generates diffs
func (sm *SnapshotManager) CompareSnapshots(oldDir, newDir string) ([]APIDiff, error) {
	log.Println("\n🔍 Comparing snapshots...")
	log.Printf("   Old: %s\n", filepath.Base(oldDir))
	log.Printf("   New: %s\n", filepath.Base(newDir))

	// Load both snapshots
	oldPages, err := sm.LoadSnapshot(oldDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load old snapshot: %w", err)
	}

	newPages, err := sm.LoadSnapshot(newDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load new snapshot: %w", err)
	}

	log.Printf("   Loaded: %d old pages, %d new pages\n", len(oldPages), len(newPages))

	var allDiffs []APIDiff

	// Compare each page
	for filename, newPage := range newPages {
		oldPage, exists := oldPages[filename]

		if !exists {
			log.Printf("   🆕 New page: %s\n", filename)
			// New page added - extract parameters
			params := extractor.ExtractAPIParameters(newPage)
			if len(params) > 0 {
				section, method := parseURLComponents(newPage.URL)
				diff := APIDiff{
					Section:     section,
					Endpoint:    newPage.URL,
					Method:      method,
					AddedParams: params,
					Breaking:    false, // New endpoints are not breaking
				}
				allDiffs = append(allDiffs, diff)
			}
			continue
		}

		// Compare content hashes first
		oldHash := utils.HashContent(oldPage.Text)
		newHash := utils.HashContent(newPage.Text)

		if oldHash == newHash {
			// No changes
			continue
		}

		// Content changed - extract and compare parameters
		oldParams := extractor.ExtractAPIParameters(oldPage)
		newParams := extractor.ExtractAPIParameters(newPage)

		section, method := parseURLComponents(newPage.URL)

		log.Printf("   📝 %s/%s: ", section, method)

		// Compare parameters
		diff := CompareParameters(section, newPage.URL, method, oldParams, newParams)

		// Check if there are meaningful changes
		if len(diff.AddedParams) > 0 || len(diff.RemovedParams) > 0 || len(diff.ModifiedParams) > 0 {
			log.Printf("Added:%d Removed:%d Modified:%d", len(diff.AddedParams), len(diff.RemovedParams), len(diff.ModifiedParams))
			if diff.Breaking {
				log.Printf(" ⚠️ BREAKING")
			}
			log.Println()
			allDiffs = append(allDiffs, diff)
		} else {
			log.Println("content changed but no parameter changes")
			// Mark as raw content change
			diff.RawContentChange = true
			// Only add to diffs if there's actually meaningful content change
			// (not just whitespace/formatting)
			if len(oldParams) > 0 || len(newParams) > 0 {
				allDiffs = append(allDiffs, diff)
			}
		}
	}

	// Check for removed pages
	for filename := range oldPages {
		if _, exists := newPages[filename]; !exists {
			log.Printf("   ❌ Removed page: %s\n", filename)
		}
	}

	log.Printf("\n✅ Comparison complete: %d endpoints with changes\n", len(allDiffs))

	return allDiffs, nil
}

// CleanupOldSnapshots removes snapshots older than the specified retention period
func (sm *SnapshotManager) CleanupOldSnapshots(retainCount int) error {
	entries, err := os.ReadDir(sm.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	// Filter directories only
	var snapshots []string
	for _, entry := range entries {
		if entry.IsDir() {
			snapshots = append(snapshots, entry.Name())
		}
	}

	// Keep only the last N snapshots
	if len(snapshots) > retainCount {
		toRemove := snapshots[:len(snapshots)-retainCount]
		for _, snapshot := range toRemove {
			path := filepath.Join(sm.BaseDir, snapshot)
			if err := os.RemoveAll(path); err != nil {
				log.Printf("⚠️  Failed to remove old snapshot %s: %v\n", snapshot, err)
			} else {
				log.Printf("🗑️  Removed old snapshot: %s\n", snapshot)
			}
		}
	}

	return nil
}

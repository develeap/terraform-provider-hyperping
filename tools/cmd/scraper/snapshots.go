package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/openapi"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/utils"
)

// SnapshotManager handles storing and retrieving OpenAPI YAML snapshots.
// Each snapshot is a single hyperping-api-<timestamp>.yaml file stored in
// a timestamped sub-directory under BaseDir.
type SnapshotManager struct {
	BaseDir string // e.g., "snapshots/"
}

// NewSnapshotManager creates a new snapshot manager.
func NewSnapshotManager(baseDir string) *SnapshotManager {
	return &SnapshotManager{BaseDir: baseDir}
}

// SaveSnapshot writes the OpenAPI spec to a timestamped directory.
// Filename: snapshots/2026-02-03_10-30-00/hyperping-api.yaml
func (sm *SnapshotManager) SaveSnapshot(timestamp time.Time, spec *openapi.Spec) error {
	snapshotDir := filepath.Join(sm.BaseDir, timestamp.Format("2006-01-02_15-04-05"))

	if err := os.MkdirAll(snapshotDir, 0750); err != nil {
		return fmt.Errorf("snapshot: create dir %s: %w", snapshotDir, err)
	}

	specPath := filepath.Join(snapshotDir, "hyperping-api.yaml")
	if err := openapi.Save(spec, specPath); err != nil {
		return err
	}

	log.Printf("📸 Snapshot saved: %s\n", snapshotDir)
	return nil
}

// GetLatestSnapshot returns the path to the most recent hyperping-api.yaml.
func (sm *SnapshotManager) GetLatestSnapshot() (string, error) {
	entries, err := os.ReadDir(sm.BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("snapshot: no snapshots exist yet")
		}
		return "", fmt.Errorf("snapshot: read dir %s: %w", sm.BaseDir, err)
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	if len(dirs) == 0 {
		return "", fmt.Errorf("snapshot: no snapshots found in %s", sm.BaseDir)
	}

	// Directories sort lexicographically by timestamp format, so last = newest.
	latest := dirs[len(dirs)-1]
	return filepath.Join(sm.BaseDir, latest, "hyperping-api.yaml"), nil
}

// CompareSnapshots loads the two most recent snapshots and compares them.
// Returns the path pair (old, new) for use with diff.Compare.
func (sm *SnapshotManager) CompareSnapshots() (oldPath, newPath string, err error) {
	entries, err := os.ReadDir(sm.BaseDir)
	if err != nil {
		return "", "", fmt.Errorf("snapshot: read dir: %w", err)
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	if len(dirs) < 2 {
		return "", "", fmt.Errorf("snapshot: need at least 2 snapshots for comparison, have %d", len(dirs))
	}

	prevDir := dirs[len(dirs)-2]
	currDir := dirs[len(dirs)-1]

	oldPath = filepath.Join(sm.BaseDir, prevDir, "hyperping-api.yaml")
	newPath = filepath.Join(sm.BaseDir, currDir, "hyperping-api.yaml")

	log.Printf("🔍 Comparing snapshots: %s → %s\n", prevDir, currDir)
	return oldPath, newPath, nil
}

// CleanupOldSnapshots removes snapshot directories beyond the retention count.
func (sm *SnapshotManager) CleanupOldSnapshots(retainCount int) error {
	entries, err := os.ReadDir(sm.BaseDir)
	if err != nil {
		return fmt.Errorf("snapshot: read dir: %w", err)
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	if len(dirs) <= retainCount {
		return nil
	}

	toRemove := dirs[:len(dirs)-retainCount]
	for _, name := range toRemove {
		path := filepath.Join(sm.BaseDir, name)
		if err := os.RemoveAll(path); err != nil {
			log.Printf("⚠️  Failed to remove old snapshot %s: %v\n", name, err)
		} else {
			log.Printf("🗑️  Removed old snapshot: %s\n", name)
		}
	}

	return nil
}

// SaveLatestOpenAPI writes an always-current copy to docs_scraped/openapi/ for CI.
func SaveLatestOpenAPI(spec *openapi.Spec, outputDir string) error {
	dir := filepath.Join(outputDir, "openapi")
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("snapshot: create openapi dir: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	path := filepath.Join(dir, "hyperping-api-"+date+".yaml")

	if err := openapi.Save(spec, path); err != nil {
		return err
	}

	// Also write a stable "latest" symlink target for tooling.
	latestPath := filepath.Join(dir, "hyperping-api-latest.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return utils.AtomicWriteFile(latestPath, data, utils.FilePermPrivate)
}

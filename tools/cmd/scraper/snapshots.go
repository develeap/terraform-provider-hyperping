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
	specPath := filepath.Join(sm.BaseDir, latest, "hyperping-api.yaml")
	if _, err := os.Stat(specPath); err != nil {
		return "", fmt.Errorf("snapshot: spec file missing in %s (cache incomplete?): %w", latest, err)
	}
	return specPath, nil
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

	if _, err := os.Stat(oldPath); err != nil {
		return "", "", fmt.Errorf("snapshot: base spec missing (cache incomplete?): %s", oldPath)
	}
	if _, err := os.Stat(newPath); err != nil {
		return "", "", fmt.Errorf("snapshot: current spec missing: %s", newPath)
	}

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

// EnumRegression records a single field whose enum shrank between two snapshots.
type EnumRegression struct {
	Path      string // OAS path, e.g. "/v1/monitors"
	Method    string // HTTP method, e.g. "POST"
	Field     string // property name, e.g. "regions"
	OldValues []string
	NewValues []string
}

// DetectEnumRegression loads the spec at prevPath and compares its enum values
// against newSpec. It returns one EnumRegression for every request-body property
// whose enum has fewer values in newSpec than in the previous snapshot.
//
// Enum shrinkage almost always indicates a degraded scrape (lazy-loaded content
// not fully rendered) rather than a genuine API change. Callers should refuse to
// save newSpec as a baseline when regressions are present.
func DetectEnumRegression(prevPath string, newSpec *openapi.Spec) ([]EnumRegression, error) {
	prevSpec, err := openapi.Load(prevPath)
	if err != nil {
		return nil, fmt.Errorf("enum regression: load prev spec %s: %w", prevPath, err)
	}

	var regressions []EnumRegression

	for path, prevItem := range prevSpec.Paths {
		newItem, ok := newSpec.Paths[path]
		if !ok {
			continue
		}
		for _, pair := range []struct {
			method string
			prev   *openapi.Operation
			curr   *openapi.Operation
		}{
			{"POST", prevItem.Post, newItem.Post},
			{"PUT", prevItem.Put, newItem.Put},
			{"PATCH", prevItem.Patch, newItem.Patch},
		} {
			if pair.prev == nil || pair.curr == nil {
				continue
			}
			if pair.prev.RequestBody == nil || pair.curr.RequestBody == nil {
				continue
			}
			prevMT, ok := pair.prev.RequestBody.Content["application/json"]
			if !ok {
				continue
			}
			newMT, ok := pair.curr.RequestBody.Content["application/json"]
			if !ok {
				continue
			}
			for field, prevSchema := range prevMT.Schema.Properties {
				if len(prevSchema.Enum) == 0 {
					continue
				}
				newSchema := newMT.Schema.Properties[field]
				if len(newSchema.Enum) < len(prevSchema.Enum) {
					regressions = append(regressions, EnumRegression{
						Path:      path,
						Method:    pair.method,
						Field:     field,
						OldValues: prevSchema.Enum,
						NewValues: newSchema.Enum,
					})
				}
			}
		}
	}

	return regressions, nil
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

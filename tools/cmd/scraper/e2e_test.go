package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/diff"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/discovery"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/openapi"
)

const docsScrapedDir = "./docs_scraped"

// loadDocsScraped reads all *.json files from dir and returns
// a map of docPath (e.g. "monitors/create") → *PageData.
func loadDocsScraped(dir string) (map[string]*extractor.PageData, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	pages := make(map[string]*extractor.PageData)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var pd extractor.PageData
		if err := json.Unmarshal(data, &pd); err != nil {
			return nil, err
		}
		// Derive docPath from the URL embedded in the JSON.
		// e.g. "https://hyperping.com/docs/api/monitors/create" → "monitors/create"
		const apiPrefix = "/docs/api/"
		idx := strings.Index(pd.URL, apiPrefix)
		if idx == -1 {
			continue
		}
		docPath := strings.TrimSuffix(pd.URL[idx+len(apiPrefix):], "/")
		if docPath != "" {
			pages[docPath] = &pd
		}
	}
	return pages, nil
}

// TestE2E_Extraction verifies that the HTML extractor produces sensible output
// for every existing docs_scraped/*.json file.
func TestE2E_Extraction(t *testing.T) {
	pages, err := loadDocsScraped(docsScrapedDir)
	if err != nil {
		t.Skipf("docs_scraped not found: %v", err)
	}
	if len(pages) == 0 {
		t.Skip("no docs_scraped files")
	}

	t.Logf("Loaded %d docs_scraped files", len(pages))

	totalWithParams := 0
	totalParams := 0
	for docPath, pd := range pages {
		params := extractor.ExtractAPIParameters(pd)
		t.Logf("  %-40s → %d params", docPath, len(params))
		if len(params) > 0 {
			totalWithParams++
			totalParams += len(params)
			for _, p := range params {
				if p.Name == "" {
					t.Errorf("%s: extracted param with empty name", docPath)
				}
			}
		}
	}

	t.Logf("Pages with params: %d/%d, total params: %d", totalWithParams, len(pages), totalParams)
	if totalWithParams == 0 {
		t.Error("expected at least some pages to yield parameters")
	}
	if totalParams == 0 {
		t.Error("expected non-zero total parameters")
	}
}

// TestE2E_OASGeneration verifies that extracted params produce a valid OpenAPI spec.
func TestE2E_OASGeneration(t *testing.T) {
	pages, err := loadDocsScraped(docsScrapedDir)
	if err != nil {
		t.Skipf("docs_scraped not found: %v", err)
	}

	apiParams := make(map[string][]extractor.APIParameter)
	for docPath, pd := range pages {
		if params := extractor.ExtractAPIParameters(pd); len(params) > 0 {
			apiParams[docPath] = params
		}
	}

	spec := openapi.Generate(apiParams, "e2e-test")
	if spec == nil {
		t.Fatal("openapi.Generate returned nil")
	}
	if spec.OpenAPI != "3.0.3" {
		t.Errorf("expected OpenAPI=3.0.3, got %q", spec.OpenAPI)
	}
	if len(spec.Paths) == 0 {
		t.Error("generated spec has no paths — check openapi/mappings.go")
	}

	t.Logf("Generated OAS spec: %d paths", len(spec.Paths))
	for path, item := range spec.Paths {
		methods := []string{}
		if item.Get != nil {
			methods = append(methods, "GET")
		}
		if item.Post != nil {
			methods = append(methods, "POST")
		}
		if item.Put != nil {
			methods = append(methods, "PUT")
		}
		if item.Delete != nil {
			methods = append(methods, "DELETE")
		}
		t.Logf("  %s [%s]", path, strings.Join(methods, ","))
	}
}

// TestE2E_SnapshotRoundtrip verifies save→load of an OAS snapshot.
func TestE2E_SnapshotRoundtrip(t *testing.T) {
	pages, err := loadDocsScraped(docsScrapedDir)
	if err != nil {
		t.Skipf("docs_scraped not found: %v", err)
	}

	apiParams := make(map[string][]extractor.APIParameter)
	for docPath, pd := range pages {
		if params := extractor.ExtractAPIParameters(pd); len(params) > 0 {
			apiParams[docPath] = params
		}
	}

	spec := openapi.Generate(apiParams, "e2e-test")

	tmpDir := t.TempDir()
	sm := NewSnapshotManager(tmpDir)

	// Save snapshot.
	ts := time.Now()
	if err := sm.SaveSnapshot(ts, spec); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	// Retrieve and load it back.
	latestPath, err := sm.GetLatestSnapshot()
	if err != nil {
		t.Fatalf("GetLatestSnapshot: %v", err)
	}

	loaded, err := openapi.Load(latestPath)
	if err != nil {
		t.Fatalf("openapi.Load: %v", err)
	}

	if loaded.OpenAPI != spec.OpenAPI {
		t.Errorf("round-trip: OpenAPI mismatch: %q vs %q", loaded.OpenAPI, spec.OpenAPI)
	}
	if len(loaded.Paths) != len(spec.Paths) {
		t.Errorf("round-trip: paths count mismatch: %d vs %d", len(loaded.Paths), len(spec.Paths))
	}
	t.Logf("Snapshot round-trip OK (%d paths)", len(loaded.Paths))
}

// TestE2E_DiffNoChange verifies that comparing a spec to itself reports no changes.
func TestE2E_DiffNoChange(t *testing.T) {
	pages, err := loadDocsScraped(docsScrapedDir)
	if err != nil {
		t.Skipf("docs_scraped not found: %v", err)
	}

	apiParams := make(map[string][]extractor.APIParameter)
	for docPath, pd := range pages {
		if params := extractor.ExtractAPIParameters(pd); len(params) > 0 {
			apiParams[docPath] = params
		}
	}

	spec := openapi.Generate(apiParams, "e2e-test")
	tmpDir := t.TempDir()
	sm := NewSnapshotManager(tmpDir)

	// Save the same snapshot twice (different timestamps to create two dirs).
	t1 := time.Now()
	if err := sm.SaveSnapshot(t1, spec); err != nil {
		t.Fatal(err)
	}
	t2 := t1.Add(time.Second)
	if err := sm.SaveSnapshot(t2, spec); err != nil {
		t.Fatal(err)
	}

	oldPath, newPath, err := sm.CompareSnapshots()
	if err != nil {
		t.Fatalf("CompareSnapshots: %v", err)
	}

	result, err := diff.Compare(oldPath, newPath)
	if err != nil {
		t.Fatalf("diff.Compare: %v", err)
	}

	if result.Breaking {
		t.Error("expected no breaking changes when comparing spec to itself")
	}
	t.Logf("Diff no-change OK (breaking=%v, hasChanges=%v)", result.Breaking, result.HasChanges)
}

// TestE2E_SitemapDiscovery fetches the live Hyperping sitemap.
// Skipped when -short is set.
func TestE2E_SitemapDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in -short mode")
	}

	urls, err := discovery.DiscoverFromSitemap("")
	if err != nil {
		t.Fatalf("DiscoverFromSitemap: %v", err)
	}
	if len(urls) == 0 {
		t.Error("expected at least one API URL from sitemap")
	}

	t.Logf("Sitemap discovery: %d API URLs found", len(urls))
	for _, u := range urls {
		t.Logf("  [%-15s] %s", u.Section, u.DocPath)
	}
}

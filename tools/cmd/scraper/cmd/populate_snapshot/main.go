// populate_snapshot is a helper that creates a snapshot from existing docs_scraped/ files.
// Used only for local testing without a browser.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/extractor"
	"github.com/develeap/terraform-provider-hyperping/tools/scraper/openapi"
)

func main() {
	docsDir := "./docs_scraped"
	snapshotBaseDir := "/tmp/e2e-snapshots"

	entries, err := os.ReadDir(docsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read docs_scraped: %v\n", err)
		os.Exit(1)
	}

	apiParams := make(map[string][]extractor.APIParameter)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(docsDir, e.Name()))
		if err != nil {
			continue
		}
		var pd extractor.PageData
		if err := json.Unmarshal(data, &pd); err != nil {
			continue
		}

		const apiPrefix = "/docs/api/"
		idx := strings.Index(pd.URL, apiPrefix)
		if idx == -1 {
			continue
		}
		docPath := strings.TrimSuffix(pd.URL[idx+len(apiPrefix):], "/")
		if docPath == "" {
			continue
		}

		params := extractor.ExtractAPIParameters(&pd)
		if len(params) > 0 {
			apiParams[docPath] = params
			fmt.Printf("  %-40s → %d params\n", docPath, len(params))
		}
	}

	spec := openapi.Generate(apiParams, time.Now().Format("2006-01-02"))
	fmt.Printf("\nGenerated spec with %d paths\n", len(spec.Paths))

	snapshotDir := filepath.Join(snapshotBaseDir, time.Now().Format("2006-01-02_15-04-05"))
	if err := os.MkdirAll(snapshotDir, 0750); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	specPath := filepath.Join(snapshotDir, "hyperping-api.yaml")
	if err := openapi.Save(spec, specPath); err != nil {
		fmt.Fprintf(os.Stderr, "save: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Snapshot saved: %s\n", specPath)
}

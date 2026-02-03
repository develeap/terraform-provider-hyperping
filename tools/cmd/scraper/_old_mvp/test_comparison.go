package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// loadPageData loads a PageData struct from a JSON file
func loadPageData(filepath string) (*PageData, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var pageData PageData
	if err := json.Unmarshal(data, &pageData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &pageData, nil
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	fmt.Println("🔍 Testing Snapshot Comparison")
	fmt.Println("===============================\n")

	snapshotMgr := NewSnapshotManager("snapshots")

	oldSnapshot := "snapshots/2026-02-03_10-53-23"
	newSnapshot := "snapshots/2026-02-03_11-05-43"

	diffs, err := snapshotMgr.CompareSnapshots(oldSnapshot, newSnapshot)
	if err != nil {
		log.Fatalf("❌ Comparison failed: %v\n", err)
	}

	if len(diffs) == 0 {
		fmt.Println("\n❌ No diffs found!")
		return
	}

	fmt.Printf("\n✅ Found %d endpoint(s) with changes\n\n", len(diffs))

	for i, diff := range diffs {
		fmt.Printf("Diff %d:\n", i+1)
		fmt.Printf("  Section: %s\n", diff.Section)
		fmt.Printf("  Method: %s\n", diff.Method)
		fmt.Printf("  Added params: %d\n", len(diff.AddedParams))
		fmt.Printf("  Removed params: %d\n", len(diff.RemovedParams))
		fmt.Printf("  Modified params: %d\n", len(diff.ModifiedParams))
		fmt.Printf("  Breaking: %v\n", diff.Breaking)

		if len(diff.AddedParams) > 0 {
			fmt.Println("  Added parameters:")
			for _, p := range diff.AddedParams {
				fmt.Printf("    - %s (%s, required=%v)\n", p.Name, p.Type, p.Required)
			}
		}
	}
}

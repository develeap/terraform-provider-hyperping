// run_diff compares two OAS YAML snapshots and prints a human-readable result.
package main

import (
	"fmt"
	"os"

	"github.com/develeap/terraform-provider-hyperping/tools/scraper/diff"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: run_diff <base.yaml> <curr.yaml>")
		os.Exit(1)
	}
	result, err := diff.Compare(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "diff failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("HasChanges: %v\n", result.HasChanges)
	fmt.Printf("Breaking:   %v\n", result.Breaking)
	if result.Summary != "" {
		fmt.Println("\nSummary:")
		fmt.Println(result.Summary)
	} else {
		fmt.Println("\nNo diff summary (specs are identical).")
	}
}

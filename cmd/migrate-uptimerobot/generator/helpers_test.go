// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite testdata/*.golden files with current output")

// goldenAssert compares got to the contents of testdata/<name>. With
// -update-golden, the file is rewritten instead. testdata/ is created on
// demand so a deleted golden regenerates rather than failing in a confusing way.
func goldenAssert(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil { //nolint:gosec // testdata only
			t.Fatalf("write %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v (run: go test ./cmd/migrate-uptimerobot/generator -update-golden)", path, err)
	}
	if got != string(want) {
		t.Errorf("output does not match %s\nrun -update-golden after intentional changes\n--- got ---\n%s\n--- want ---\n%s", name, got, string(want))
	}
}

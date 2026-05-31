// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	hyperping "github.com/develeap/hyperping-go"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

// TestGenerateImportScript_RejectsMaliciousUUID is defense in depth for the
// pingdom migration import-script emitter. The createdResources map carries
// UUIDs returned by the Hyperping API; if those values were ever
// attacker-influenced, the previous use of fmt.Sprintf("%q", uuid) would not
// neutralize bash metacharacters and the malicious payload would execute when
// the script ran.
func TestGenerateImportScript_RejectsMaliciousUUID(t *testing.T) {
	checks := []pingdom.Check{
		{ID: 1, Name: "evil-one", Type: "http"},
		{ID: 2, Name: "evil-two", Type: "http"},
		{ID: 3, Name: "good", Type: "http"},
	}
	results := []converter.ConversionResult{
		{Supported: true, Monitor: &hyperping.CreateMonitorRequest{Name: "evil-one", URL: "https://e1.example.com", Protocol: "http"}},
		{Supported: true, Monitor: &hyperping.CreateMonitorRequest{Name: "evil-two", URL: "https://e2.example.com", Protocol: "http"}},
		{Supported: true, Monitor: &hyperping.CreateMonitorRequest{Name: "good", URL: "https://g.example.com", Protocol: "http"}},
	}
	createdResources := map[int]string{
		1: "$(rm -rf $HOME)",
		2: "'; rm -rf $HOME; '",
		3: "mon_safe_123",
	}

	g := NewImportGenerator("")
	out := g.GenerateImportScript(checks, results, createdResources)

	forbidden := []string{"$(rm", "'; rm"}
	for _, f := range forbidden {
		if strings.Contains(out, f) {
			t.Errorf("script contains forbidden substring %q\n---\n%s\n---", f, out)
		}
	}
	if !strings.Contains(out, "mon_safe_123") {
		t.Errorf("safe UUID was rejected; script:\n%s", out)
	}

	// Same defense for the GenerateImportCommands path.
	out2 := g.GenerateImportCommands(checks, results, createdResources)
	for _, f := range forbidden {
		if strings.Contains(out2, f) {
			t.Errorf("commands output contains forbidden substring %q\n---\n%s\n---", f, out2)
		}
	}
}

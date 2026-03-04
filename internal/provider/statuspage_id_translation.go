// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// warnUnresolvedNumericUUIDs checks for services that have numeric UUIDs
// instead of mon_xxx format. This indicates the API stored a numeric ID
// (legacy drift from older provider versions or API behavior). Since we
// store raw API values in state, the plan engine will detect these as diffs
// against the config's mon_xxx UUIDs and trigger an apply to fix them.
func warnUnresolvedNumericUUIDs(sp *client.StatusPage, diags *diag.Diagnostics) {
	if sp == nil {
		return
	}

	var drifted []string
	for _, section := range sp.Sections {
		collectDriftedUUIDs(section.Services, &drifted)
	}

	if len(drifted) > 0 {
		diags.AddWarning(
			"Status page services have numeric UUIDs (drift detected)",
			fmt.Sprintf(
				"Found %d service(s) with numeric UUIDs instead of mon_xxx format: %s. "+
					"The API returned numeric IDs instead of monitor UUIDs. "+
					"Run 'terraform apply' to fix — the provider will send the correct UUIDs from your config.",
				len(drifted), strings.Join(drifted, ", ")),
		)
	}
}

// collectDriftedUUIDs recursively collects service UUIDs that are numeric strings.
func collectDriftedUUIDs(services []client.StatusPageService, drifted *[]string) {
	for _, svc := range services {
		if svc.UUID != "" && isNumericString(svc.UUID) {
			*drifted = append(*drifted, svc.UUID)
		}
		if len(svc.Services) > 0 {
			collectDriftedUUIDs(svc.Services, drifted)
		}
	}
}

// isNumericString returns true if s contains only digits.
func isNumericString(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

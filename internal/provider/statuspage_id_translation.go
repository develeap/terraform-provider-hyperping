// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// buildMonitorIDToUUIDMap fetches all monitors and builds a lookup map from
// v1 numeric IDs (e.g. "115746") to UUID strings (e.g. "mon_abc123").
// Used on the read path to translate numeric IDs in API responses back to UUIDs.
// On error, adds a warning diagnostic and returns an empty map (graceful fallback).
func buildMonitorIDToUUIDMap(ctx context.Context, apiClient client.MonitorAPI, diags *diag.Diagnostics) map[string]string {
	idToUUID := make(map[string]string)

	monitors, err := apiClient.ListMonitors(ctx)
	if err != nil {
		diags.AddWarning(
			"Unable to fetch monitor list for ID translation",
			fmt.Sprintf("Status page services may show numeric IDs instead of UUIDs: %s", err),
		)
		return idToUUID
	}

	for _, mon := range monitors {
		if mon.UUID != "" && mon.ID != 0 {
			idToUUID[strconv.Itoa(mon.ID)] = mon.UUID
		}
	}

	return idToUUID
}

// translateStatusPageToUUIDs walks the StatusPage response and replaces numeric
// ID strings with UUIDs in-place. This ensures Terraform always stores UUIDs
// regardless of what the API returned.
func translateStatusPageToUUIDs(sp *client.StatusPage, idToUUID map[string]string) {
	if sp == nil || len(idToUUID) == 0 {
		return
	}

	for i := range sp.Sections {
		translateSectionServicesToUUIDs(sp.Sections[i].Services, idToUUID)
	}
}

// translateSectionServicesToUUIDs translates services within a section from numeric IDs to UUIDs.
func translateSectionServicesToUUIDs(services []client.StatusPageService, idToUUID map[string]string) {
	for i := range services {
		svc := &services[i]

		// Check if UUID field is a numeric string that needs translation
		if svc.UUID != "" {
			if uuid, ok := idToUUID[svc.UUID]; ok {
				svc.UUID = uuid
			}
		}

		// Check if ID field is a numeric value that needs translation
		idStr := serviceIDToNumericString(svc.ID)
		if idStr != "" {
			if uuid, ok := idToUUID[idStr]; ok {
				svc.UUID = uuid
			}
		}

		// Recurse into nested services
		if len(svc.Services) > 0 {
			translateSectionServicesToUUIDs(svc.Services, idToUUID)
		}
	}
}

// serviceIDToNumericString extracts a numeric string from the flexible ID field.
func serviceIDToNumericString(id interface{}) string {
	switch v := id.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v)
	case string:
		if _, err := strconv.Atoi(v); err == nil {
			return v
		}
	}
	return ""
}

// warnUnresolvedNumericUUIDs checks for services that still have numeric UUIDs
// after translation (i.e., the API returned a numeric ID that couldn't be mapped
// to a mon_xxx UUID). This indicates legacy drift from older provider versions
// that translated UUIDs to numeric IDs on write. The next terraform apply will
// fix these by sending the correct mon_xxx UUIDs from the config.
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
			"Status page services have numeric UUIDs (legacy drift detected)",
			fmt.Sprintf(
				"Found %d service(s) with numeric UUIDs instead of mon_xxx format: %s. "+
					"This was caused by an older provider version that translated UUIDs to numeric IDs. "+
					"Run 'terraform apply' to fix — the provider will send the correct UUIDs from your config.",
				len(drifted), strings.Join(drifted, ", ")),
		)
	}
}

// collectDriftedUUIDs recursively collects service UUIDs that are still numeric strings.
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

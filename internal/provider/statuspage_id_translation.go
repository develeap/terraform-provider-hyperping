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

// monitorIDMaps holds bidirectional lookup maps between monitor UUIDs and numeric IDs.
type monitorIDMaps struct {
	uuidToNumericID map[string]string // mon_xxx -> "117896"
	numericIDToUUID map[string]string // "117896" -> mon_xxx
}

// buildMonitorIDMaps fetches all monitors and builds bidirectional lookup maps.
// Takes a function argument for testability (call sites pass r.client.ListMonitors).
func buildMonitorIDMaps(
	ctx context.Context,
	listMonitors func(context.Context) ([]client.Monitor, error),
) (*monitorIDMaps, error) {
	monitors, err := listMonitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}
	maps := &monitorIDMaps{
		uuidToNumericID: make(map[string]string, len(monitors)),
		numericIDToUUID: make(map[string]string, len(monitors)),
	}
	for _, m := range monitors {
		numericID := strconv.Itoa(m.ID)
		maps.uuidToNumericID[m.UUID] = numericID
		maps.numericIDToUUID[numericID] = m.UUID
	}
	return maps, nil
}

// translateSectionsUUIDsToNumericIDs translates mon_xxx UUIDs to numeric IDs
// in all services within sections. Adds error diagnostic if any UUID is unresolvable.
func translateSectionsUUIDsToNumericIDs(
	sections []client.CreateStatusPageSection,
	uuidToID map[string]string,
	diags *diag.Diagnostics,
) {
	var unresolved []string
	for i := range sections {
		translateCreateServicesToNumericIDs(sections[i].Services, uuidToID, &unresolved)
	}
	if len(unresolved) > 0 {
		diags.AddError(
			"Cannot resolve monitor UUIDs to numeric IDs",
			fmt.Sprintf("%d monitor UUID(s) could not be resolved: %s. "+
				"Ensure all referenced monitors exist.", len(unresolved), strings.Join(unresolved, ", ")),
		)
	}
}

// translateCreateServicesToNumericIDs translates UUIDs in CreateStatusPageService slice.
// Top-level: MonitorUUID field. Nested (inside groups): UUID field.
func translateCreateServicesToNumericIDs(
	services []client.CreateStatusPageService,
	uuidToID map[string]string,
	unresolved *[]string,
) {
	for i := range services {
		svc := &services[i]
		// Top-level service: translate MonitorUUID
		if svc.MonitorUUID != nil && *svc.MonitorUUID != "" {
			if numericID, ok := uuidToID[*svc.MonitorUUID]; ok {
				svc.MonitorUUID = &numericID
			} else {
				*unresolved = append(*unresolved, *svc.MonitorUUID)
			}
		}
		// Nested service: translate UUID
		if svc.UUID != nil && *svc.UUID != "" {
			if numericID, ok := uuidToID[*svc.UUID]; ok {
				svc.UUID = &numericID
			} else {
				*unresolved = append(*unresolved, *svc.UUID)
			}
		}
		// Recurse into nested services (groups)
		if len(svc.Services) > 0 {
			translateCreateServicesToNumericIDs(svc.Services, uuidToID, unresolved)
		}
	}
}

// translateResponseNumericIDsToUUIDs translates numeric IDs back to mon_xxx
// in the API response. Warns if any numeric ID can't be resolved.
func translateResponseNumericIDsToUUIDs(
	sp *client.StatusPage,
	idToUUID map[string]string,
	diags *diag.Diagnostics,
) {
	if sp == nil {
		return
	}
	var unresolved []string
	for i := range sp.Sections {
		translateServicesToUUIDs(sp.Sections[i].Services, idToUUID, &unresolved)
	}
	if len(unresolved) > 0 {
		diags.AddWarning(
			"Some service numeric IDs could not be resolved to UUIDs",
			fmt.Sprintf("%d service(s) have numeric IDs that don't match any monitor: %s. "+
				"These monitors may have been deleted.",
				len(unresolved), strings.Join(unresolved, ", ")),
		)
	}
}

// translateServicesToUUIDs translates StatusPageService UUIDs from numeric to mon_xxx.
func translateServicesToUUIDs(
	services []client.StatusPageService,
	idToUUID map[string]string,
	unresolved *[]string,
) {
	for i := range services {
		svc := &services[i]
		if svc.UUID != "" && isNumericString(svc.UUID) {
			if uuid, ok := idToUUID[svc.UUID]; ok {
				svc.UUID = uuid
			} else {
				*unresolved = append(*unresolved, svc.UUID)
			}
		}
		// Recurse into nested services
		if len(svc.Services) > 0 {
			translateServicesToUUIDs(svc.Services, idToUUID, unresolved)
		}
	}
}

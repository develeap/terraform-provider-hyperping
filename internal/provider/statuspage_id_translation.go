// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// buildMonitorIDMaps fetches all monitors and builds bidirectional lookup maps
// between UUID strings (e.g. "mon_abc123") and v1 numeric IDs (e.g. 115746).
// On error, adds a warning diagnostic and returns empty maps (graceful fallback).
func buildMonitorIDMaps(ctx context.Context, apiClient client.HyperpingAPI, diags *diag.Diagnostics) (uuidToID map[string]int, idToUUID map[string]string) {
	uuidToID = make(map[string]int)
	idToUUID = make(map[string]string)

	monitors, err := apiClient.ListMonitors(ctx)
	if err != nil {
		diags.AddWarning(
			"Unable to fetch monitor list for ID translation",
			fmt.Sprintf("Status page services will use UUIDs as-is (renderer may not resolve status): %s", err),
		)
		return uuidToID, idToUUID
	}

	for _, mon := range monitors {
		if mon.UUID != "" && mon.ID != 0 {
			uuidToID[mon.UUID] = mon.ID
			idToUUID[strconv.Itoa(mon.ID)] = mon.UUID
		}
	}

	return uuidToID, idToUUID
}

// translateSectionsToNumericIDs walks sections and replaces UUID strings with
// numeric ID strings for renderer compatibility. Creates new slices (immutable).
// Unknown UUIDs are left as-is (graceful fallback).
func translateSectionsToNumericIDs(sections []client.CreateStatusPageSection, uuidToID map[string]int) []client.CreateStatusPageSection {
	if len(uuidToID) == 0 {
		return sections
	}

	translated := make([]client.CreateStatusPageSection, len(sections))
	for i, section := range sections {
		translated[i] = client.CreateStatusPageSection{
			Name:     section.Name,
			IsSplit:  section.IsSplit,
			Services: translateServicesToNumericIDs(section.Services, uuidToID),
		}
	}

	return translated
}

// translateServicesToNumericIDs translates service UUIDs to numeric IDs.
func translateServicesToNumericIDs(services []client.CreateStatusPageService, uuidToID map[string]int) []client.CreateStatusPageService {
	if len(services) == 0 {
		return services
	}

	translated := make([]client.CreateStatusPageService, len(services))
	for i, svc := range services {
		translated[i] = translateServiceToNumericID(svc, uuidToID)
	}

	return translated
}

// translateServiceToNumericID translates a single service's UUID to numeric ID.
func translateServiceToNumericID(svc client.CreateStatusPageService, uuidToID map[string]int) client.CreateStatusPageService {
	result := client.CreateStatusPageService{
		NameShown:         svc.NameShown,
		Name:              svc.Name,
		ShowUptime:        svc.ShowUptime,
		ShowResponseTimes: svc.ShowResponseTimes,
		IsGroup:           svc.IsGroup,
	}

	// Top-level services use MonitorUUID
	if svc.MonitorUUID != nil {
		if numID, ok := uuidToID[*svc.MonitorUUID]; ok {
			idStr := strconv.Itoa(numID)
			result.MonitorUUID = &idStr
		} else {
			result.MonitorUUID = svc.MonitorUUID
		}
	}

	// Nested services use UUID
	if svc.UUID != nil {
		if numID, ok := uuidToID[*svc.UUID]; ok {
			idStr := strconv.Itoa(numID)
			result.UUID = &idStr
		} else {
			result.UUID = svc.UUID
		}
	}

	// Recurse into nested services for groups
	if len(svc.Services) > 0 {
		result.Services = translateServicesToNumericIDs(svc.Services, uuidToID)
	}

	return result
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

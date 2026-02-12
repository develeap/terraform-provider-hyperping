// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import "fmt"

// Test helper functions for generating test data.
// These are used by performance tests to generate large datasets.

// MonitorUUID generates a test monitor UUID.
func MonitorUUID(i int) string {
	return fmt.Sprintf("mon_test_%06d", i)
}

// MonitorName generates a test monitor name based on index and category.
func MonitorName(i, category int) string {
	categories := []string{"API", "WEB", "DATABASE", "SERVICE"}
	envs := []string{"PROD", "STAGING", "DEV", "TEST"}

	env := envs[i%len(envs)]
	cat := categories[category%len(categories)]

	return fmt.Sprintf("[%s]-%s-Service-%d", env, cat, i)
}

// IncidentID generates a test incident ID.
func IncidentID(i int) string {
	return fmt.Sprintf("inc_test_%06d", i)
}

// IncidentTitle generates a test incident title.
func IncidentTitle(i int) string {
	types := []string{"API Issue", "Database Outage", "Network Problem", "Service Degradation"}
	return fmt.Sprintf("%s #%d", types[i%len(types)], i)
}

// MaintenanceID generates a test maintenance window ID.
func MaintenanceID(i int) string {
	return fmt.Sprintf("maint_test_%06d", i)
}

// MaintenanceTitle generates a test maintenance window title.
func MaintenanceTitle(i int) string {
	types := []string{"Database Upgrade", "Server Maintenance", "Network Update", "Security Patch"}
	return fmt.Sprintf("%s - Window %d", types[i%len(types)], i)
}

// HealthcheckID generates a test healthcheck ID.
func HealthcheckID(i int) string {
	return fmt.Sprintf("hc_test_%06d", i)
}

// HealthcheckName generates a test healthcheck name.
func HealthcheckName(i int) string {
	return fmt.Sprintf("Healthcheck-%d", i)
}

// OutageID generates a test outage ID.
func OutageID(i int) string {
	return fmt.Sprintf("outage_test_%06d", i)
}

// StatusPageID generates a test status page ID.
func StatusPageID(i int) string {
	return fmt.Sprintf("sp_test_%06d", i)
}

// StatusPageName generates a test status page name.
func StatusPageName(i int) string {
	return fmt.Sprintf("Status Page %d", i)
}

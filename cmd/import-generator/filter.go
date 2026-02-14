// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"regexp"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// FilterConfig holds filtering criteria for resources.
type FilterConfig struct {
	NamePattern    *regexp.Regexp
	ExcludePattern *regexp.Regexp
	ResourceType   string
}

// NewFilterConfig creates a new filter configuration from command-line arguments.
func NewFilterConfig(namePattern, excludePattern, resourceType string) (*FilterConfig, error) {
	fc := &FilterConfig{
		ResourceType: resourceType,
	}

	if namePattern != "" {
		re, err := regexp.Compile(namePattern)
		if err != nil {
			return nil, err
		}
		fc.NamePattern = re
	}

	if excludePattern != "" {
		re, err := regexp.Compile(excludePattern)
		if err != nil {
			return nil, err
		}
		fc.ExcludePattern = re
	}

	return fc, nil
}

// IsEmpty returns true if no filters are configured.
func (fc *FilterConfig) IsEmpty() bool {
	return fc.NamePattern == nil && fc.ExcludePattern == nil && fc.ResourceType == ""
}

// ShouldIncludeResourceType returns true if the given resource type should be processed.
func (fc *FilterConfig) ShouldIncludeResourceType(resourceType string) bool {
	if fc.ResourceType == "" {
		return true
	}
	return fc.ResourceType == resourceType
}

// FilterMonitors applies filters to monitor resources.
func (fc *FilterConfig) FilterMonitors(monitors []client.Monitor) []client.Monitor {
	if fc.IsEmpty() || !fc.ShouldIncludeResourceType("hyperping_monitor") {
		if !fc.ShouldIncludeResourceType("hyperping_monitor") {
			return nil
		}
		return monitors
	}

	filtered := make([]client.Monitor, 0, len(monitors))
	for _, m := range monitors {
		if fc.matchesName(m.Name) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// FilterHealthchecks applies filters to healthcheck resources.
func (fc *FilterConfig) FilterHealthchecks(healthchecks []client.Healthcheck) []client.Healthcheck {
	if fc.IsEmpty() || !fc.ShouldIncludeResourceType("hyperping_healthcheck") {
		if !fc.ShouldIncludeResourceType("hyperping_healthcheck") {
			return nil
		}
		return healthchecks
	}

	filtered := make([]client.Healthcheck, 0, len(healthchecks))
	for _, h := range healthchecks {
		if fc.matchesName(h.Name) {
			filtered = append(filtered, h)
		}
	}
	return filtered
}

// FilterStatusPages applies filters to status page resources.
func (fc *FilterConfig) FilterStatusPages(pages []client.StatusPage) []client.StatusPage {
	if fc.IsEmpty() || !fc.ShouldIncludeResourceType("hyperping_statuspage") {
		if !fc.ShouldIncludeResourceType("hyperping_statuspage") {
			return nil
		}
		return pages
	}

	filtered := make([]client.StatusPage, 0, len(pages))
	for _, p := range pages {
		if fc.matchesName(p.Name) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// FilterIncidents applies filters to incident resources.
func (fc *FilterConfig) FilterIncidents(incidents []client.Incident) []client.Incident {
	if fc.IsEmpty() || !fc.ShouldIncludeResourceType("hyperping_incident") {
		if !fc.ShouldIncludeResourceType("hyperping_incident") {
			return nil
		}
		return incidents
	}

	filtered := make([]client.Incident, 0, len(incidents))
	for _, i := range incidents {
		if fc.matchesName(i.Title.En) {
			filtered = append(filtered, i)
		}
	}
	return filtered
}

// FilterMaintenance applies filters to maintenance resources.
func (fc *FilterConfig) FilterMaintenance(maintenance []client.Maintenance) []client.Maintenance {
	if fc.IsEmpty() || !fc.ShouldIncludeResourceType("hyperping_maintenance") {
		if !fc.ShouldIncludeResourceType("hyperping_maintenance") {
			return nil
		}
		return maintenance
	}

	filtered := make([]client.Maintenance, 0, len(maintenance))
	for _, m := range maintenance {
		titleText := m.Title.En
		if titleText == "" {
			titleText = m.Name
		}
		if fc.matchesName(titleText) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// FilterOutages applies filters to outage resources.
func (fc *FilterConfig) FilterOutages(outages []client.Outage) []client.Outage {
	if fc.IsEmpty() || !fc.ShouldIncludeResourceType("hyperping_outage") {
		if !fc.ShouldIncludeResourceType("hyperping_outage") {
			return nil
		}
		return outages
	}

	filtered := make([]client.Outage, 0, len(outages))
	for _, o := range outages {
		if fc.matchesName(o.Monitor.Name) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

// matchesName returns true if the name matches the filter criteria.
func (fc *FilterConfig) matchesName(name string) bool {
	// Check exclude pattern first
	if fc.ExcludePattern != nil && fc.ExcludePattern.MatchString(name) {
		return false
	}

	// If name pattern is set, must match
	if fc.NamePattern != nil {
		return fc.NamePattern.MatchString(name)
	}

	// No name pattern, include by default
	return true
}

// Summary returns a human-readable summary of the filter configuration.
func (fc *FilterConfig) Summary() string {
	if fc.IsEmpty() {
		return "No filters applied"
	}

	parts := []string{}

	if fc.NamePattern != nil {
		parts = append(parts, "Name pattern: "+fc.NamePattern.String())
	}

	if fc.ExcludePattern != nil {
		parts = append(parts, "Exclude pattern: "+fc.ExcludePattern.String())
	}

	if fc.ResourceType != "" {
		parts = append(parts, "Resource type: "+fc.ResourceType)
	}

	return strings.Join(parts, ", ")
}

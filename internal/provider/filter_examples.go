// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

// This file contains example usage patterns for the filter framework.
// These examples demonstrate how to integrate filtering into data sources.

/*
EXAMPLE 1: Basic Name Filtering for Monitors Data Source

In your data source schema:

	func (d *MonitorsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
		resp.Schema = schema.Schema{
			Attributes: map[string]schema.Attribute{
				"filter": MonitorFilterSchema(),
				"monitors": schema.ListNestedAttribute{
					Computed: true,
					// ... nested attributes
				},
			},
		}
	}

In your data source model:

	type MonitorsDataSourceModel struct {
		Filter   *MonitorFilterModel `tfsdk:"filter"`
		Monitors []MonitorDataModel  `tfsdk:"monitors"`
	}

In your Read function:

	func (d *MonitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
		var config MonitorsDataSourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

		// Fetch all monitors from API
		allMonitors, err := d.client.ListMonitors(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error listing monitors", err.Error())
			return
		}

		// Apply client-side filtering
		var filteredMonitors []MonitorDataModel
		for _, monitor := range allMonitors {
			if shouldIncludeMonitor(monitor, config.Filter) {
				filteredMonitors = append(filteredMonitors, mapMonitorToModel(monitor))
			}
		}

		config.Monitors = filteredMonitors
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	}

	func shouldIncludeMonitor(monitor client.Monitor, filter *MonitorFilterModel) bool {
		if filter == nil {
			return true // No filter = include all
		}

		// Use ApplyAllFilters for clean short-circuit evaluation
		return ApplyAllFilters(
			func() bool {
				match, _ := MatchesNameRegex(monitor.Name, filter.NameRegex)
				return match
			},
			func() bool {
				return MatchesExactCaseInsensitive(monitor.Protocol, filter.Protocol)
			},
			func() bool {
				return MatchesBool(monitor.Paused, filter.Paused)
			},
		)
	}

Terraform usage:

	data "hyperping_monitors" "production_http" {
		filter = {
			name_regex = "\\[PROD\\]-.*"
			protocol   = "http"
			paused     = false
		}
	}

EXAMPLE 2: Incident Filtering with Multiple Criteria

In your Read function:

	func shouldIncludeIncident(incident client.Incident, filter *IncidentFilterModel) bool {
		if filter == nil {
			return true
		}

		return ApplyAllFilters(
			func() bool {
				match, err := MatchesNameRegex(incident.Title, filter.NameRegex)
				if err != nil {
					// Log error but don't fail - just exclude this incident
					return false
				}
				return match
			},
			func() bool {
				return MatchesExact(incident.Status, filter.Status)
			},
			func() bool {
				return MatchesExact(incident.Severity, filter.Severity)
			},
		)
	}

Terraform usage:

	data "hyperping_incidents" "critical_active" {
		filter = {
			status   = "investigating"
			severity = "critical"
		}
	}

EXAMPLE 3: Maintenance Windows with Date Range (future enhancement)

If we add date range filtering in the future:

	type MaintenanceFilterModel struct {
		NameRegex      types.String `tfsdk:"name_regex"`
		Status         types.String `tfsdk:"status"`
		StartAfter     types.String `tfsdk:"start_after"`  // ISO 8601
		StartBefore    types.String `tfsdk:"start_before"` // ISO 8601
	}

	func shouldIncludeMaintenanceWindow(mw client.MaintenanceWindow, filter *MaintenanceFilterModel) bool {
		if filter == nil {
			return true
		}

		// Parse dates if needed
		var startAfter, startBefore *time.Time
		if !isNullOrUnknown(filter.StartAfter) {
			t, _ := time.Parse(time.RFC3339, filter.StartAfter.ValueString())
			startAfter = &t
		}
		if !isNullOrUnknown(filter.StartBefore) {
			t, _ := time.Parse(time.RFC3339, filter.StartBefore.ValueString())
			startBefore = &t
		}

		return ApplyAllFilters(
			func() bool {
				match, _ := MatchesNameRegex(mw.Title, filter.NameRegex)
				return match
			},
			func() bool {
				return MatchesExact(mw.Status, filter.Status)
			},
			func() bool {
				if startAfter == nil {
					return true
				}
				return mw.ScheduledStart.After(*startAfter)
			},
			func() bool {
				if startBefore == nil {
					return true
				}
				return mw.ScheduledStart.Before(*startBefore)
			},
		)
	}

EXAMPLE 4: Outages Filtering by Monitor UUID

	func shouldIncludeOutage(outage client.Outage, filter *OutageFilterModel) bool {
		if filter == nil {
			return true
		}

		return ApplyAllFilters(
			func() bool {
				match, _ := MatchesNameRegex(outage.MonitorName, filter.NameRegex)
				return match
			},
			func() bool {
				return MatchesExact(outage.MonitorUUID, filter.MonitorUUID)
			},
		)
	}

Terraform usage:

	# Get all outages for a specific monitor
	data "hyperping_outages" "api_outages" {
		filter = {
			monitor_uuid = hyperping_monitor.api.id
		}
	}

EXAMPLE 5: Status Pages with Hostname Filtering

	func shouldIncludeStatusPage(sp client.StatusPage, filter *StatusPageFilterModel) bool {
		if filter == nil {
			return true
		}

		return ApplyAllFilters(
			func() bool {
				match, _ := MatchesNameRegex(sp.Name, filter.NameRegex)
				return match
			},
			func() bool {
				return MatchesExact(sp.Hostname, filter.Hostname)
			},
		)
	}

EXAMPLE 6: Error Handling Pattern

When regex compilation fails, you should handle it gracefully:

	func (d *MonitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
		var config MonitorsDataSourceModel
		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

		// Validate regex pattern before fetching from API
		if config.Filter != nil && !isNullOrUnknown(config.Filter.NameRegex) {
			_, err := regexp.Compile(config.Filter.NameRegex.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid Filter Regex",
					fmt.Sprintf("The name_regex pattern is invalid: %s", err.Error()),
				)
				return
			}
		}

		// Fetch and filter...
	}

PERFORMANCE NOTES:

1. Regex Compilation Caching (for future optimization):

	type MonitorsDataSource struct {
		client      client.MonitorAPI
		regexCache  map[string]*regexp.Regexp
		cacheMutex  sync.RWMutex
	}

	func (d *MonitorsDataSource) getCompiledRegex(pattern string) (*regexp.Regexp, error) {
		d.cacheMutex.RLock()
		if re, ok := d.regexCache[pattern]; ok {
			d.cacheMutex.RUnlock()
			return re, nil
		}
		d.cacheMutex.RUnlock()

		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}

		d.cacheMutex.Lock()
		if d.regexCache == nil {
			d.regexCache = make(map[string]*regexp.Regexp)
		}
		d.regexCache[pattern] = re
		d.cacheMutex.Unlock()

		return re, nil
	}

2. Early Exit Pattern:

	// ApplyAllFilters already implements short-circuit evaluation
	// It returns false immediately on first non-match

3. Large Dataset Optimization:

	// For very large datasets, consider parallel filtering
	func filterMonitorsConcurrent(monitors []client.Monitor, filter *MonitorFilterModel) []MonitorDataModel {
		var wg sync.WaitGroup
		resultChan := make(chan MonitorDataModel, len(monitors))

		for _, monitor := range monitors {
			wg.Add(1)
			go func(m client.Monitor) {
				defer wg.Done()
				if shouldIncludeMonitor(m, filter) {
					resultChan <- mapMonitorToModel(m)
				}
			}(monitor)
		}

		go func() {
			wg.Wait()
			close(resultChan)
		}()

		var results []MonitorDataModel
		for result := range resultChan {
			results = append(results, result)
		}
		return results
	}

TESTING PATTERN:

	func TestMonitorsDataSourceFiltering(t *testing.T) {
		tests := []struct {
			name     string
			monitors []client.Monitor
			filter   *MonitorFilterModel
			expected []string // expected monitor names
		}{
			{
				name: "no filter returns all",
				monitors: []client.Monitor{
					{Name: "api-1"},
					{Name: "api-2"},
				},
				filter:   nil,
				expected: []string{"api-1", "api-2"},
			},
			{
				name: "regex filter matches subset",
				monitors: []client.Monitor{
					{Name: "[PROD]-api"},
					{Name: "[DEV]-api"},
				},
				filter: &MonitorFilterModel{
					NameRegex: types.StringValue("\\[PROD\\]-.*"),
				},
				expected: []string{"[PROD]-api"},
			},
			{
				name: "protocol filter",
				monitors: []client.Monitor{
					{Name: "api-1", Protocol: "http"},
					{Name: "api-2", Protocol: "https"},
				},
				filter: &MonitorFilterModel{
					Protocol: types.StringValue("http"),
				},
				expected: []string{"api-1"},
			},
			{
				name: "combined filters",
				monitors: []client.Monitor{
					{Name: "[PROD]-api", Protocol: "http", Paused: false},
					{Name: "[PROD]-db", Protocol: "tcp", Paused: false},
					{Name: "[PROD]-cache", Protocol: "http", Paused: true},
				},
				filter: &MonitorFilterModel{
					NameRegex: types.StringValue("\\[PROD\\]-.*"),
					Protocol:  types.StringValue("http"),
					Paused:    types.BoolValue(false),
				},
				expected: []string{"[PROD]-api"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var results []string
				for _, monitor := range tt.monitors {
					if shouldIncludeMonitor(monitor, tt.filter) {
						results = append(results, monitor.Name)
					}
				}

				if len(results) != len(tt.expected) {
					t.Errorf("got %v results, want %v", len(results), len(tt.expected))
				}

				for i, expected := range tt.expected {
					if results[i] != expected {
						t.Errorf("result[%d] = %q, want %q", i, results[i], expected)
					}
				}
			})
		}
	}
*/

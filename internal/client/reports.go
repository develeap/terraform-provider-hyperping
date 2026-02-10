// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
	"net/url"
)

// reportsBasePath uses the exported constant for consistency.
var reportsBasePath = ReportsBasePath

// GetMonitorReport returns the report for a specific monitor.
// Optional from/to parameters can be provided for date filtering.
func (c *Client) GetMonitorReport(ctx context.Context, uuid string, from, to string) (*MonitorReport, error) {
	if err := ValidateResourceID(uuid); err != nil {
		return nil, fmt.Errorf("GetMonitorReport: %w", err)
	}
	path := fmt.Sprintf("%s/%s", reportsBasePath, uuid)

	// Add query parameters if provided
	if from != "" || to != "" {
		params := url.Values{}
		if from != "" {
			params.Set("from", from)
		}
		if to != "" {
			params.Set("to", to)
		}
		path = path + "?" + params.Encode()
	}

	var report MonitorReport
	if err := c.doRequest(ctx, "GET", path, nil, &report); err != nil {
		return nil, fmt.Errorf("failed to get monitor report %s: %w", uuid, err)
	}
	return &report, nil
}

// ListMonitorReports returns reports for all monitors.
// Optional from/to parameters can be provided for date filtering.
// API returns wrapped response: {"period": {...}, "monitors": [...]}
func (c *Client) ListMonitorReports(ctx context.Context, from, to string) ([]MonitorReport, error) {
	path := reportsBasePath

	// Add query parameters if provided
	if from != "" || to != "" {
		params := url.Values{}
		if from != "" {
			params.Set("from", from)
		}
		if to != "" {
			params.Set("to", to)
		}
		path = path + "?" + params.Encode()
	}

	var response ListMonitorReportsResponse
	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list monitor reports: %w", err)
	}
	return response.Monitors, nil
}

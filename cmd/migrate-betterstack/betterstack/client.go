// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package betterstack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://betteruptime.com/api/v2"
	defaultTimeout = 30 * time.Second
)

// Client is a Better Stack API client.
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new Better Stack API client.
func NewClient(apiToken string) *Client {
	return &Client{
		baseURL:  defaultBaseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Monitor represents a Better Stack monitor.
type Monitor struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Attributes MonitorAttributes `json:"attributes"`
}

// MonitorAttributes contains monitor configuration.
type MonitorAttributes struct {
	PronouncableName    string          `json:"pronounceable_name"`
	URL                 string          `json:"url"`
	MonitorType         string          `json:"monitor_type"`
	CheckFrequency      int             `json:"check_frequency"`
	RequestTimeout      int             `json:"request_timeout"`
	RequestMethod       string          `json:"request_method"`
	RequestHeaders      []RequestHeader `json:"request_headers"`
	RequestBody         string          `json:"request_body"`
	ExpectedStatusCodes []int           `json:"expected_status_codes"`
	FollowRedirects     bool            `json:"follow_redirects"`
	Paused              bool            `json:"paused"`
	MonitorGroupID      int             `json:"monitor_group_id"`
	Regions             []string        `json:"regions"`
	Port                int             `json:"port,omitempty"`
}

// RequestHeader represents an HTTP request header.
type RequestHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Heartbeat represents a Better Stack heartbeat.
type Heartbeat struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"`
	Attributes HeartbeatAttributes `json:"attributes"`
}

// HeartbeatAttributes contains heartbeat configuration.
type HeartbeatAttributes struct {
	Name   string `json:"name"`
	Period int    `json:"period"`
	Grace  int    `json:"grace"`
	Paused bool   `json:"paused"`
}

// MonitorsResponse is the API response for listing monitors.
type MonitorsResponse struct {
	Data       []Monitor  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// HeartbeatsResponse is the API response for listing heartbeats.
type HeartbeatsResponse struct {
	Data       []Heartbeat `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination contains pagination metadata.
type Pagination struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
}

// FetchMonitors retrieves all monitors from Better Stack.
func (c *Client) FetchMonitors(ctx context.Context) ([]Monitor, error) {
	var allMonitors []Monitor
	page := 1

	for {
		url := fmt.Sprintf("%s/monitors?page=%d&per_page=100", c.baseURL, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiToken)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var result MonitorsResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}

		allMonitors = append(allMonitors, result.Data...)

		// Check if there are more pages
		if result.Pagination.Next == "" {
			break
		}
		page++
	}

	return allMonitors, nil
}

// FetchHeartbeats retrieves all heartbeats from Better Stack.
func (c *Client) FetchHeartbeats(ctx context.Context) ([]Heartbeat, error) {
	var allHeartbeats []Heartbeat
	page := 1

	for {
		url := fmt.Sprintf("%s/heartbeats?page=%d&per_page=100", c.baseURL, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiToken)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var result HeartbeatsResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}

		allHeartbeats = append(allHeartbeats, result.Data...)

		// Check if there are more pages
		if result.Pagination.Next == "" {
			break
		}
		page++
	}

	return allHeartbeats, nil
}

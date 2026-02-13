// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package uptimerobot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL = "https://api.uptimerobot.com/v2"
)

// Client is an UptimeRobot API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new UptimeRobot API client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Monitor represents an UptimeRobot monitor.
type Monitor struct {
	ID            int               `json:"id"`
	FriendlyName  string            `json:"friendly_name"`
	URL           string            `json:"url"`
	Type          int               `json:"type"` // 1=HTTP, 2=Keyword, 3=Ping, 4=Port, 5=Heartbeat
	SubType       *int              `json:"sub_type,omitempty"`
	KeywordType   *int              `json:"keyword_type,omitempty"` // 1=exists, 2=not exists
	KeywordValue  *string           `json:"keyword_value,omitempty"`
	HTTPMethod    *int              `json:"http_method,omitempty"` // 1=GET, 2=POST, 3=PUT, 4=PATCH, 5=DELETE, 6=HEAD
	Port          *int              `json:"port,omitempty"`
	Interval      int               `json:"interval"` // Check interval in seconds
	Timeout       *int              `json:"timeout,omitempty"`
	Status        int               `json:"status"` // 0=paused, 1=not checked yet, 2=up, 8=seems down, 9=down
	AlertContacts []AlertContactRef `json:"alert_contacts,omitempty"`
}

// AlertContactRef represents a reference to an alert contact in a monitor.
type AlertContactRef struct {
	ID         string `json:"id"`
	Type       int    `json:"type"`
	Value      string `json:"value"`
	Threshold  int    `json:"threshold"`
	Recurrence int    `json:"recurrence"`
}

// AlertContact represents an UptimeRobot alert contact.
type AlertContact struct {
	ID           string `json:"id"`
	FriendlyName string `json:"friendly_name"`
	Type         int    `json:"type"` // 2=Email, 3=SMS, 4=Webhook, 11=Slack, 14=PagerDuty, etc.
	Value        string `json:"value"`
	Status       int    `json:"status"`
}

// GetMonitorsResponse represents the response from getMonitors endpoint.
type GetMonitorsResponse struct {
	Stat     string    `json:"stat"`
	Monitors []Monitor `json:"monitors"`
	Error    *APIError `json:"error,omitempty"`
}

// GetAlertContactsResponse represents the response from getAlertContacts endpoint.
type GetAlertContactsResponse struct {
	Stat          string         `json:"stat"`
	AlertContacts []AlertContact `json:"alert_contacts"`
	Error         *APIError      `json:"error,omitempty"`
}

// APIError represents an UptimeRobot API error.
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// GetMonitors fetches all monitors from UptimeRobot.
func (c *Client) GetMonitors(ctx context.Context) ([]Monitor, error) {
	payload := map[string]interface{}{
		"api_key":        c.apiKey,
		"format":         "json",
		"logs":           0,
		"alert_contacts": 1,
		"response_times": 0,
	}

	resp, err := c.doRequest(ctx, "getMonitors", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var result GetMonitorsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if result.Stat != "ok" {
		if result.Error != nil {
			return nil, fmt.Errorf("API error: %s - %s", result.Error.Type, result.Error.Message)
		}
		return nil, fmt.Errorf("API returned status: %s", result.Stat)
	}

	return result.Monitors, nil
}

// GetAlertContacts fetches all alert contacts from UptimeRobot.
func (c *Client) GetAlertContacts(ctx context.Context) ([]AlertContact, error) {
	payload := map[string]interface{}{
		"api_key": c.apiKey,
		"format":  "json",
	}

	resp, err := c.doRequest(ctx, "getAlertContacts", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var result GetAlertContactsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if result.Stat != "ok" {
		if result.Error != nil {
			return nil, fmt.Errorf("API error: %s - %s", result.Error.Type, result.Error.Message)
		}
		return nil, fmt.Errorf("API returned status: %s", result.Stat)
	}

	return result.AlertContacts, nil
}

// doRequest performs an HTTP POST request to the UptimeRobot API.
func (c *Client) doRequest(ctx context.Context, endpoint string, payload map[string]interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", baseURL, endpoint)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}

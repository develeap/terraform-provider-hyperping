// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package pingdom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.pingdom.com/api/3.1"

// Client represents a Pingdom API client.
type Client struct {
	apiToken   string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Pingdom API client.
func NewClient(apiToken string, options ...Option) *Client {
	c := &Client{
		apiToken: apiToken,
		baseURL:  defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithBaseURL sets the base URL for the Pingdom API.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// Check represents a Pingdom check.
type Check struct {
	ID                       int               `json:"id"`
	Name                     string            `json:"name"`
	Type                     string            `json:"type"` // http, tcp, ping, dns, udp, smtp, pop3, imap
	Hostname                 string            `json:"hostname"`
	URL                      string            `json:"url,omitempty"`
	Encryption               bool              `json:"encryption"`
	Port                     int               `json:"port,omitempty"`
	Resolution               int               `json:"resolution"` // Check frequency in minutes
	Paused                   bool              `json:"paused"`
	Tags                     []Tag             `json:"tags"`
	ProbeFilters             []string          `json:"probe_filters,omitempty"`
	RequestHeaders           map[string]string `json:"requestheaders,omitempty"`
	PostData                 string            `json:"postdata,omitempty"`
	ShouldContain            string            `json:"shouldcontain,omitempty"`
	ShouldNotContain         string            `json:"shouldnotcontain,omitempty"`
	VerifyCertificate        bool              `json:"verify_certificate"`
	SSLDownDaysBefore        int               `json:"ssl_down_days_before,omitempty"`
	SendNotificationWhenDown int               `json:"sendnotificationwhendown,omitempty"`
	NotifyAgainEvery         int               `json:"notifyagainevery,omitempty"`
	NotifyWhenBackup         bool              `json:"notifywhenbackup"`
	CustomMessage            string            `json:"custom_message,omitempty"`
	IntegrationIDs           []int             `json:"integrationids,omitempty"`
	TeamIDs                  []int             `json:"teamids,omitempty"`
	UserIDs                  []int             `json:"userids,omitempty"`
	Auth                     string            `json:"auth,omitempty"`
	AdditionalURLs           []string          `json:"additional_urls,omitempty"`
	StringToExpect           string            `json:"stringtoexpect,omitempty"`
	StringToSend             string            `json:"stringtosend,omitempty"`
	ExpectedIP               string            `json:"expectedip,omitempty"`
	NameServer               string            `json:"nameserver,omitempty"`
}

// Tag represents a Pingdom tag.
type Tag struct {
	Name string `json:"name"`
	Type string `json:"type"` // u (user-defined) or a (auto)
}

// ChecksResponse represents the response from the /checks endpoint.
type ChecksResponse struct {
	Checks []Check `json:"checks"`
}

// CheckDetailResponse represents the response from the /checks/{id} endpoint.
type CheckDetailResponse struct {
	Check Check `json:"check"`
}

// ListChecks fetches all checks from Pingdom.
func (c *Client) ListChecks(ctx context.Context) ([]Check, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/checks", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req) //nolint:gosec // G704: baseURL is operator-configured, not user-tainted input
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ChecksResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return response.Checks, nil
}

// GetCheck fetches detailed information about a specific check.
func (c *Client) GetCheck(ctx context.Context, checkID int) (*Check, error) {
	url := fmt.Sprintf("%s/checks/%d", c.baseURL, checkID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req) //nolint:gosec // G704: baseURL is operator-configured, not user-tainted input
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response CheckDetailResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &response.Check, nil
}

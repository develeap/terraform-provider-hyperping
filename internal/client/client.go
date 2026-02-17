// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package client provides a Go client for the Hyperping API.
// This package is intentionally separate from the Terraform provider
// to allow independent testing and potential reuse.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sony/gobreaker"
)

const (
	// DefaultBaseURL is the default Hyperping API base URL.
	DefaultBaseURL = "https://api.hyperping.io"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retry attempts.
	DefaultMaxRetries = 3

	// DefaultRetryWaitMin is the minimum wait time between retries.
	DefaultRetryWaitMin = 1 * time.Second //nolint:revive // Min/Max pairing is clearer than removing suffix

	// DefaultRetryWaitMax is the maximum wait time between retries.
	DefaultRetryWaitMax = 30 * time.Second

	// maxResourceIDLength is the maximum allowed length for resource IDs (VULN-013).
	maxResourceIDLength = 128

	// maxResponseBodyBytes is the maximum response body size to read into memory (VULN-014).
	// 10 MB is generous for JSON API responses while preventing OOM from malicious servers.
	maxResponseBodyBytes = 10 * 1024 * 1024

	// maxUserAgentLength is the maximum allowed User-Agent string length (VULN-010).
	maxUserAgentLength = 256
)

// retryableStatusCodes are HTTP status codes that should trigger a retry.
var retryableStatusCodes = map[int]bool{
	429: true, // Too Many Requests
	500: true, // Internal Server Error
	502: true, // Bad Gateway
	503: true, // Service Unavailable
	504: true, // Gateway Timeout
}

// Logger is an optional interface for logging HTTP requests and responses.
// This allows the client to be used with any logging framework.
type Logger interface {
	// Debug logs a debug-level message with optional key-value pairs.
	Debug(ctx context.Context, msg string, fields map[string]interface{})
}

// Metrics is an optional interface for collecting operational metrics.
// This allows integration with Prometheus, CloudWatch, Datadog, etc.
type Metrics interface {
	// RecordAPICall records an API call with method, path, status code, and duration.
	RecordAPICall(ctx context.Context, method, path string, statusCode int, durationMs int64)
	// RecordRetry records a retry attempt.
	RecordRetry(ctx context.Context, method, path string, attempt int)
	// RecordCircuitBreakerState records circuit breaker state changes.
	RecordCircuitBreakerState(ctx context.Context, state string)
}

// Client is the Hyperping API client.
type Client struct {
	baseURL        string
	httpClient     *http.Client
	maxRetries     int
	retryWaitMin   time.Duration
	retryWaitMax   time.Duration
	logger         Logger
	metrics        Metrics
	version        string
	userAgent      string
	circuitBreaker *gobreaker.CircuitBreaker
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// NewClient creates a new Hyperping API client.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &http.Transport{
				// Connection pooling to prevent exhaustion under load (VULN-021)
				MaxIdleConnsPerHost: 10,               // Allow 10 idle connections per host
				MaxConnsPerHost:     20,               // Limit 20 total connections per host
				IdleConnTimeout:     90 * time.Second, // Reuse idle connections for 90s
			},
		},
		maxRetries:   DefaultMaxRetries,
		retryWaitMin: DefaultRetryWaitMin,
		retryWaitMax: DefaultRetryWaitMax,
		version:      "dev", // Default version, can be overridden with WithVersion
	}

	for _, opt := range opts {
		opt(c)
	}

	// Build transport chain after all options are applied (VULN-009, VULN-011).
	// Chain: baseTransport → TLS enforcement → auth injection.
	// Auth is injected via RoundTripper to avoid persistent string allocations.
	// TLS is enforced on the transport layer, even for custom HTTP clients.
	baseTransport := c.httpClient.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	c.httpClient.Transport = buildTransportChain([]byte(apiKey), baseTransport, c.baseURL)

	// Initialize circuit breaker to prevent cascading failures
	c.circuitBreaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "hyperping-api",
		MaxRequests: 3,                // Max concurrent requests in half-open state
		Interval:    60 * time.Second, // Rolling window for failure counting
		Timeout:     30 * time.Second, // Time to wait before attempting recovery (half-open)
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Early return before division to prevent divide-by-zero if Requests is 0 (VULN-020)
			if counts.Requests < 3 {
				return false
			}
			return float64(counts.TotalFailures)/float64(counts.Requests) >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			ctx := context.Background()
			if c.logger != nil {
				c.logger.Debug(ctx, "circuit breaker state change", map[string]interface{}{
					"name": name,
					"from": from.String(),
					"to":   to.String(),
				})
			}
			if c.metrics != nil {
				c.metrics.RecordCircuitBreakerState(ctx, to.String())
			}
		},
	})

	// Build User-Agent after all options are applied
	c.userAgent = buildUserAgent(c.version)

	return c
}

// WithBaseURL sets a custom base URL for the client.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryWait sets the minimum and maximum retry wait times.
func WithRetryWait(minWait, maxWait time.Duration) Option {
	return func(c *Client) {
		c.retryWaitMin = minWait
		c.retryWaitMax = maxWait
	}
}

// WithLogger sets a logger for HTTP request/response logging.
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithMetrics sets a metrics collector for operational observability.
func WithMetrics(metrics Metrics) Option {
	return func(c *Client) {
		c.metrics = metrics
	}
}

// WithVersion sets the provider version for the User-Agent header.
func WithVersion(version string) Option {
	return func(c *Client) {
		c.version = version
	}
}

// resourceIDPattern validates Hyperping resource IDs.
// Allowed formats: type_alphanumeric (e.g., mon_abc123, out_xyz789, inc_001)
// Also allows plain alphanumeric IDs without prefix for forward compatibility.
var resourceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateResourceID validates that a resource ID is safe for use in URL paths.
// Prevents path traversal (../), URL parameter injection (?key=val), and
// SSRF via URL authority override (@evil.com).
// Exported for use in provider ImportState validation (VULN-015).
func ValidateResourceID(id string) error {
	if id == "" {
		return fmt.Errorf("resource ID must not be empty")
	}

	// Reject oversized IDs to prevent memory exhaustion in URL construction (VULN-013)
	if len(id) > maxResourceIDLength {
		return fmt.Errorf("invalid resource ID: length %d exceeds maximum of %d", len(id), maxResourceIDLength)
	}

	// Reject path traversal sequences
	if strings.Contains(id, "..") || strings.Contains(id, "/") {
		return fmt.Errorf("invalid resource ID %q: path traversal not allowed", id)
	}

	// Reject URL parameter injection
	if strings.ContainsAny(id, "?#@&=") {
		return fmt.Errorf("invalid resource ID %q: URL metacharacters not allowed", id)
	}

	// Enforce allowed character pattern
	if !resourceIDPattern.MatchString(id) {
		return fmt.Errorf("invalid resource ID %q: must match [a-zA-Z0-9_-]", id)
	}

	return nil
}

// buildUserAgent constructs the User-Agent string for API requests.
// Format: terraform-provider-hyperping/VERSION (go/VERSION; OS/ARCH) [TF_APPEND_USER_AGENT]
//
// Example: terraform-provider-hyperping/1.0.0 (go1.24.11; linux/amd64)
// With TF_APPEND_USER_AGENT: terraform-provider-hyperping/1.0.0 (go1.24.11; linux/amd64) custom-module/2.0
func buildUserAgent(version string) string {
	// Base User-Agent: terraform-provider-hyperping/VERSION (go/VERSION; OS/ARCH)
	userAgent := fmt.Sprintf("terraform-provider-hyperping/%s (%s; %s/%s)",
		version,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)

	// Support TF_APPEND_USER_AGENT environment variable (VULN-010).
	// Sanitize to prevent HTTP header injection via CR/LF characters
	// and control character smuggling.
	if appendUA := os.Getenv("TF_APPEND_USER_AGENT"); appendUA != "" {
		sanitized := sanitizeUserAgent(appendUA)
		if sanitized != "" {
			userAgent = fmt.Sprintf("%s %s", userAgent, sanitized)
		}
	}

	// Cap total User-Agent length to prevent oversized headers
	if len(userAgent) > maxUserAgentLength {
		userAgent = userAgent[:maxUserAgentLength]
	}

	return userAgent
}

// sanitizeUserAgent strips control characters and non-printable bytes from a
// User-Agent component to prevent HTTP header injection (VULN-010).
func sanitizeUserAgent(input string) string {
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		// Allow printable ASCII (0x20-0x7E) and common Unicode letters/digits
		if r >= 0x20 && r != 0x7F {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// logDebug logs a debug message if a logger is configured.
func (c *Client) logDebug(ctx context.Context, msg string, fields map[string]interface{}) {
	if c.logger != nil {
		c.logger.Debug(ctx, msg, fields)
	}
}

// doRequest performs an HTTP request with retry logic wrapped in a circuit breaker.
func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	// Wrap request in circuit breaker to prevent cascading failures
	// If circuit breaker is not initialized (e.g., in tests), execute directly
	if c.circuitBreaker != nil {
		_, err := c.circuitBreaker.Execute(func() (interface{}, error) {
			return nil, c.doRequestWithRetry(ctx, method, path, body, result)
		})
		return err
	}
	return c.doRequestWithRetry(ctx, method, path, body, result)
}

// buildHTTPRequest constructs an HTTP request with all required headers set.
// Authorization is injected by authTransport; this sets content-type, accept, and user-agent.
func (c *Client) buildHTTPRequest(ctx context.Context, method, path string, jsonBody []byte) (*http.Request, error) {
	var bodyReader io.Reader
	if jsonBody != nil {
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Authorization header is injected by authTransport (VULN-009).
	req.Header.Set(HeaderContentType, ContentTypeJSON)
	req.Header.Set(HeaderAccept, ContentTypeJSON)
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
}

// readResponseBody reads and validates the response body with a size cap to
// prevent OOM from malicious servers (VULN-014). It always closes resp.Body.
func readResponseBody(resp *http.Response) ([]byte, error) {
	// Read maxResponseBodyBytes+1 to detect truncation.
	limitedReader := io.LimitReader(resp.Body, int64(maxResponseBodyBytes)+1)
	body, err := io.ReadAll(limitedReader)
	if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
		// Only surface close error when read succeeded.
		return nil, fmt.Errorf("failed to close response body: %w", closeErr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if len(body) > maxResponseBodyBytes {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", maxResponseBodyBytes)
	}
	return body, nil
}

// shouldRetryOnError records metrics and sleeps before deciding whether to
// retry after a transport-level error (non-HTTP failure). Returns true when
// a retry should be attempted.
func (c *Client) shouldRetryOnError(ctx context.Context, method, path string, attempt int) bool {
	if attempt >= c.maxRetries {
		return false
	}
	if c.metrics != nil {
		c.metrics.RecordRetry(ctx, method, path, attempt+1)
	}
	c.sleep(ctx, c.calculateBackoff(attempt, 0))
	return true
}

// attemptResult holds the outcome of a single HTTP attempt for use in the retry loop.
type attemptResult struct {
	// retry indicates this attempt should be retried.
	retry bool
	// err is the terminal error to return (nil on success).
	err error
	// lastErr is the transient error stored for retry reporting.
	lastErr error
}

// processHTTPResponse handles a successful HTTP transport response:
// reads the body, records metrics/logs, decodes success responses, and
// decides whether to retry on server errors. It never retries transport errors.
func (c *Client) processHTTPResponse(
	ctx context.Context,
	method, path string,
	resp *http.Response,
	duration time.Duration,
	attempt int,
	result interface{},
) attemptResult {
	respBody, err := readResponseBody(resp)
	if err != nil {
		return attemptResult{err: err}
	}

	c.logDebug(ctx, "received API response", map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": resp.StatusCode,
		"attempt":     attempt + 1,
		"duration_ms": duration.Milliseconds(),
	})

	if c.metrics != nil {
		c.metrics.RecordAPICall(ctx, method, path, resp.StatusCode, duration.Milliseconds())
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return attemptResult{err: decodeResult(respBody, result)}
	}

	return c.handleErrorResponse(ctx, method, path, resp, respBody, attempt)
}

// decodeResult unmarshals the response body into result when both are non-empty.
func decodeResult(body []byte, result interface{}) error {
	if result == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

// handleErrorResponse decides whether to retry an HTTP-level error response.
func (c *Client) handleErrorResponse(
	ctx context.Context,
	method, path string,
	resp *http.Response,
	body []byte,
	attempt int,
) attemptResult {
	apiErr := c.parseErrorResponse(resp.StatusCode, resp.Header, body)

	if !retryableStatusCodes[resp.StatusCode] || attempt >= c.maxRetries {
		return attemptResult{err: apiErr}
	}

	c.logDebug(ctx, "retrying request", map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": resp.StatusCode,
		"attempt":     attempt + 1,
		"max_retries": c.maxRetries,
	})
	if c.metrics != nil {
		c.metrics.RecordRetry(ctx, method, path, attempt+1)
	}
	c.sleep(ctx, c.calculateBackoff(attempt, apiErr.RetryAfter))
	return attemptResult{retry: true, lastErr: apiErr}
}

// marshalBody marshals the request body to JSON, returning nil for nil input.
func marshalBody(body interface{}) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return jsonBody, nil
}

// doRequestWithRetry performs an HTTP request with retry logic (internal implementation).
func (c *Client) doRequestWithRetry(ctx context.Context, method, path string, body, result interface{}) error {
	jsonBody, err := marshalBody(body)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := c.buildHTTPRequest(ctx, method, path, jsonBody)
		if err != nil {
			return err
		}

		c.logDebug(ctx, "sending API request", map[string]interface{}{
			"method":  method,
			"path":    path,
			"attempt": attempt + 1,
		})

		startTime := time.Now()
		resp, err := c.httpClient.Do(req)
		duration := time.Since(startTime)

		if err != nil {
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			c.logDebug(ctx, "request failed", map[string]interface{}{
				"method":      method,
				"path":        path,
				"error":       err.Error(),
				"attempt":     attempt + 1,
				"duration_ms": duration.Milliseconds(),
			})
			lastErr = fmt.Errorf("request failed: %w", err)
			if c.shouldRetryOnError(ctx, method, path, attempt) {
				continue
			}
			return lastErr
		}

		ar := c.processHTTPResponse(ctx, method, path, resp, duration, attempt, result)
		if ar.retry {
			lastErr = ar.lastErr
			continue
		}
		return ar.err
	}

	return lastErr
}

// parseErrorResponse parses an error response from the API.
func (c *Client) parseErrorResponse(statusCode int, headers http.Header, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
	}

	// Parse Retry-After header for rate limiting (429) responses
	if statusCode == http.StatusTooManyRequests {
		apiErr.RetryAfter = parseRetryAfter(headers.Get("Retry-After"))
	}

	// Try to parse error body
	var errResp struct {
		Error   string             `json:"error"`
		Message string             `json:"message"`
		Details []ValidationDetail `json:"details"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Error != "" {
			apiErr.Message = errResp.Error
		}
		if errResp.Message != "" && errResp.Message != errResp.Error {
			apiErr.Message = errResp.Message
		}
		apiErr.Details = errResp.Details
	}

	return apiErr
}

// parseRetryAfter parses the Retry-After header value.
// Supports two formats:
//  1. Integer seconds: "120" means wait 120 seconds
//  2. HTTP-date format: "Wed, 21 Oct 2015 07:28:00 GMT"
//
// Returns 0 if the header is missing, invalid, or in the past.
func parseRetryAfter(retryAfter string) int {
	if retryAfter == "" {
		return 0
	}

	// Try parsing as integer (seconds)
	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil {
		if seconds > 0 {
			return seconds
		}
		return 0
	}

	// Try parsing as HTTP-date (RFC1123, RFC850, or ANSI C formats)
	parsedTime, err := http.ParseTime(retryAfter)
	if err != nil {
		return 0
	}

	// Calculate seconds until the specified time
	seconds := int(time.Until(parsedTime).Seconds())
	if seconds <= 0 {
		return 0
	}

	return seconds
}

// calculateBackoff calculates the wait time for a retry attempt.
func (c *Client) calculateBackoff(attempt int, retryAfter int) time.Duration {
	// If we have a Retry-After header, use it
	if retryAfter > 0 {
		wait := time.Duration(retryAfter) * time.Second
		if wait > c.retryWaitMax {
			return c.retryWaitMax
		}
		return wait
	}

	// Cap attempt to prevent integer overflow (max 10 retries = 2^10 = 1024x multiplier)
	if attempt < 0 {
		attempt = 0
	} else if attempt > 10 {
		attempt = 10
	}

	// Exponential backoff: min * 2^attempt
	wait := c.retryWaitMin * (1 << attempt) //nolint:gosec // attempt is bounded above
	if wait > c.retryWaitMax {
		wait = c.retryWaitMax
	}

	// Add ±25% jitter to prevent timing-based information leakage (VULN-006)
	jitter := time.Duration(rand.IntN(int(wait/2))) - wait/4 // #nosec G404 -- Non-cryptographic jitter for backoff timing
	wait += jitter
	if wait < c.retryWaitMin {
		wait = c.retryWaitMin
	}
	return wait
}

// sleep waits for the specified duration, respecting context cancellation.
func (c *Client) sleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}

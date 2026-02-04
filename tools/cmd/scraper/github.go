package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// GitHubClient handles creating issues in GitHub
type GitHubClient struct {
	Token      string
	Owner      string // e.g., "develeap"
	Repo       string // e.g., "terraform-provider-hyperping"
	HTTPClient *http.Client
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(token, owner, repo string) *GitHubClient {
	return &GitHubClient{
		Token:      token,
		Owner:      owner,
		Repo:       repo,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GitHubIssue represents a GitHub issue request
type GitHubIssue struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
}

// CreateIssue creates a new GitHub issue with the diff report
func (gc *GitHubClient) CreateIssue(report DiffReport, snapshotURL string) error {
	// Generate title
	title := fmt.Sprintf("[API Change] %s", report.Summary)

	// Generate body
	body := gc.formatIssueBody(report, snapshotURL)

	// Determine labels
	labels := gc.determineLabels(report)

	// Create issue
	issue := GitHubIssue{
		Title:  title,
		Body:   body,
		Labels: labels,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", gc.Owner, gc.Repo)

	jsonData, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("failed to marshal issue: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gc.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response to get issue number
	var issueResp struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(respBody, &issueResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("\n✅ GitHub Issue Created: #%d\n", issueResp.Number)
	fmt.Printf("   URL: %s\n", issueResp.HTMLURL)

	return nil
}

// formatIssueBody formats the diff report as a GitHub issue body
func (gc *GitHubClient) formatIssueBody(report DiffReport, snapshotURL string) string {
	var body strings.Builder

	body.WriteString("## Summary\n\n")
	body.WriteString(fmt.Sprintf("**Detected:** %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05 MST")))
	body.WriteString(fmt.Sprintf("**Changes:** %s\n\n", report.Summary))

	if report.Breaking {
		body.WriteString("⚠️ **WARNING: This includes breaking changes that require immediate attention!**\n\n")
	}

	body.WriteString("---\n\n")

	// Group diffs by section
	sections := make(map[string][]APIDiff)
	for _, diff := range report.APIDiffs {
		sections[diff.Section] = append(sections[diff.Section], diff)
	}

	if len(sections) == 0 {
		body.WriteString("✅ No parameter-level changes detected.\n\n")
		body.WriteString("Content changes were detected, but no API parameters were added, removed, or modified.\n")
		return body.String()
	}

	// Write each section
	for section, diffs := range sections {
		body.WriteString(fmt.Sprintf("## %s API\n\n", strings.Title(section)))

		for _, diff := range diffs {
			// Endpoint header
			endpointName := diff.Method
			if endpointName == "" {
				endpointName = "Overview"
			}
			body.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(endpointName)))

			if diff.Breaking {
				body.WriteString("⚠️ **BREAKING CHANGE**\n\n")
			}

			// Added parameters
			if len(diff.AddedParams) > 0 {
				body.WriteString("#### 🆕 Added Parameters\n\n")
				for _, p := range diff.AddedParams {
					required := "optional"
					if p.Required {
						required = "**required**"
					}
					body.WriteString(fmt.Sprintf("- **`%s`** (%s, %s)\n", p.Name, p.Type, required))
					if p.Description != "" {
						body.WriteString(fmt.Sprintf("  - %s\n", p.Description))
					}
				}
				body.WriteString("\n")
			}

			// Removed parameters
			if len(diff.RemovedParams) > 0 {
				body.WriteString("#### ❌ Removed Parameters\n\n")
				for _, p := range diff.RemovedParams {
					required := "optional"
					if p.Required {
						required = "**required**"
					}
					body.WriteString(fmt.Sprintf("- **`%s`** (%s, %s)\n", p.Name, p.Type, required))
				}
				body.WriteString("\n")
			}

			// Modified parameters
			if len(diff.ModifiedParams) > 0 {
				body.WriteString("#### 📝 Modified Parameters\n\n")
				for _, change := range diff.ModifiedParams {
					body.WriteString(fmt.Sprintf("- **`%s`**\n", change.Name))

					if change.OldType != change.NewType {
						body.WriteString(fmt.Sprintf("  - Type: `%s` → `%s`\n", change.OldType, change.NewType))
					}

					if change.OldRequired != change.NewRequired {
						oldReq := "optional"
						newReq := "optional"
						if change.OldRequired {
							oldReq = "required"
						}
						if change.NewRequired {
							newReq = "required"
						}
						body.WriteString(fmt.Sprintf("  - Status: `%s` → `%s`", oldReq, newReq))
						if !change.OldRequired && change.NewRequired {
							body.WriteString(" ⚠️ **BREAKING**")
						}
						body.WriteString("\n")
					}
				}
				body.WriteString("\n")
			}

			body.WriteString("---\n\n")
		}
	}

	// Action items
	body.WriteString("## Action Items\n\n")
	body.WriteString("- [ ] Review all changes\n")
	body.WriteString("- [ ] Update Terraform provider schema\n")
	body.WriteString("- [ ] Update validation logic\n")
	body.WriteString("- [ ] Update tests\n")
	body.WriteString("- [ ] Update documentation\n")

	if report.Breaking {
		body.WriteString("- [ ] ⚠️ Plan migration strategy for breaking changes\n")
		body.WriteString("- [ ] Communicate breaking changes to users\n")
	}

	// Footer
	body.WriteString("\n---\n\n")
	if snapshotURL != "" {
		body.WriteString(fmt.Sprintf("📸 [View API Snapshot](%s)\n\n", snapshotURL))
	}
	body.WriteString("🤖 Auto-detected by API scraper\n")

	return body.String()
}

// determineLabels generates appropriate labels for the issue
func (gc *GitHubClient) determineLabels(report DiffReport) []string {
	labels := []string{"api-change", "automated"}

	// Breaking vs non-breaking
	if report.Breaking {
		labels = append(labels, "breaking-change")
	} else {
		labels = append(labels, "non-breaking")
	}

	// Per-API section labels
	sections := make(map[string]bool)
	for _, diff := range report.APIDiffs {
		sections[diff.Section] = true
	}

	for section := range sections {
		labels = append(labels, section+"-api")
	}

	return labels
}

// CreateLabelsIfNeeded ensures all required labels exist in the repository
func (gc *GitHubClient) CreateLabelsIfNeeded() error {
	requiredLabels := RequiredLabels

	// Get existing labels
	existingLabels, err := gc.getExistingLabels()
	if err != nil {
		return fmt.Errorf("failed to get existing labels: %w", err)
	}

	// Create missing labels
	for name, color := range requiredLabels {
		if _, exists := existingLabels[name]; !exists {
			if err := gc.createLabel(name, color); err != nil {
				fmt.Printf("⚠️  Failed to create label %s: %v\n", name, err)
			} else {
				fmt.Printf("✅ Created label: %s\n", name)
			}
		}
	}

	return nil
}

// getExistingLabels fetches all labels from the repository
func (gc *GitHubClient) getExistingLabels() (map[string]bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/labels", gc.Owner, gc.Repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+gc.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var labels []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&labels); err != nil {
		return nil, err
	}

	existing := make(map[string]bool)
	for _, label := range labels {
		existing[label.Name] = true
	}

	return existing, nil
}

// createLabel creates a new label in the repository
func (gc *GitHubClient) createLabel(name, color string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/labels", gc.Owner, gc.Repo)

	data := map[string]string{
		"name":  name,
		"color": color,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+gc.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// LoadGitHubConfig loads GitHub configuration from environment
func LoadGitHubConfig() (*GitHubClient, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Validate GitHub token format for security
	// Valid formats: ghp_ (personal), gho_ (OAuth), ghu_ (user), ghs_ (server), ghr_ (refresh)
	if !isValidGitHubToken(token) {
		return nil, fmt.Errorf("invalid GITHUB_TOKEN format - must start with ghp_, gho_, ghu_, ghs_, or ghr_ followed by alphanumeric characters")
	}

	// Parse owner/repo from GITHUB_REPOSITORY env var (set by GitHub Actions)
	// Format: "owner/repo"
	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		// Fallback to manual config
		owner := os.Getenv("GITHUB_OWNER")
		repoName := os.Getenv("GITHUB_REPO")
		if owner == "" || repoName == "" {
			return nil, fmt.Errorf("GITHUB_REPOSITORY or GITHUB_OWNER/GITHUB_REPO not set")
		}
		return NewGitHubClient(token, owner, repoName), nil
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid GITHUB_REPOSITORY format: %s", repo)
	}

	return NewGitHubClient(token, parts[0], parts[1]), nil
}

// isValidGitHubToken validates the GitHub token format
func isValidGitHubToken(token string) bool {
	// GitHub token format: prefix (4 chars) + underscore + 36+ alphanumeric chars
	// Valid prefixes: ghp, gho, ghu, ghs, ghr
	if len(token) < 40 {
		return false
	}

	validPrefixes := []string{"ghp_", "gho_", "ghu_", "ghs_", "ghr_"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(token, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return false
	}

	// Check that characters after prefix are alphanumeric or underscore
	tokenBody := token[4:] // Skip prefix
	for _, ch := range tokenBody {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}

	return true
}

// CreateOrUpdateCoverageIssue creates a new issue or updates existing one for coverage gaps
func (gc *GitHubClient) CreateOrUpdateCoverageIssue(title, body string, labels []string) error {
	// First, search for existing issue with matching title prefix
	existingIssue, err := gc.findCoverageIssue(title)
	if err != nil {
		// Continue with creation if search fails
		fmt.Printf("   ⚠️  Failed to search for existing issues: %v\n", err)
	}

	if existingIssue != nil {
		// Update existing issue (title and body)
		return gc.updateIssue(existingIssue.Number, title, body)
	}

	// Create new issue
	issue := GitHubIssue{
		Title:  title,
		Body:   body,
		Labels: labels,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", gc.Owner, gc.Repo)

	jsonData, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("failed to marshal issue: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gc.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var issueResp struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(respBody, &issueResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("   ✅ Created issue #%d: %s\n", issueResp.Number, issueResp.HTMLURL)
	return nil
}

// ExistingIssue represents a found GitHub issue
type ExistingIssue struct {
	Number int
	Title  string
}

// findCoverageIssue searches for an existing coverage gap issue
func (gc *GitHubClient) findCoverageIssue(title string) (*ExistingIssue, error) {
	// Extract resource name from title for search
	// Title format: "[Coverage Gap] hyperping_monitor: X fields need attention"
	searchTitle := title
	if idx := strings.Index(title, ":"); idx > 0 {
		searchTitle = title[:idx] // Just search for "[Coverage Gap] hyperping_monitor"
	}

	// Search open issues with coverage-gap label
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=open&labels=coverage-gap", gc.Owner, gc.Repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+gc.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var issues []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	// Find matching issue
	for _, issue := range issues {
		if strings.HasPrefix(issue.Title, searchTitle) {
			return &ExistingIssue{
				Number: issue.Number,
				Title:  issue.Title,
			}, nil
		}
	}

	return nil, nil
}

// updateIssue updates an existing issue's title and body
func (gc *GitHubClient) updateIssue(issueNumber int, title, body string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", gc.Owner, gc.Repo, issueNumber)

	data := map[string]string{
		"title": title,
		"body":  body,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+gc.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := gc.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("   ✅ Updated existing issue #%d\n", issueNumber)
	return nil
}

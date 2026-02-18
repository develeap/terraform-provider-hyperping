// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package notify wraps google/go-github for GitHub issue notifications.
// It replaces the custom github.go (605 lines) with a thin adapter (~40 lines).
package notify

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// Config holds credentials and repository coordinates for GitHub notifications.
type Config struct {
	Token string
	Owner string // e.g., "develeap"
	Repo  string // e.g., "terraform-provider-hyperping"
}

// Client wraps the GitHub API client.
type Client struct {
	gh  *github.Client
	cfg Config
}

// NewClient creates an authenticated GitHub client.
func NewClient(ctx context.Context, cfg Config) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.Token})
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		gh:  github.NewClient(tc),
		cfg: cfg,
	}
}

// NotifyAPIChange creates or updates a GitHub issue reporting API drift.
// If an open issue with the same title already exists, its body is replaced.
func (c *Client) NotifyAPIChange(ctx context.Context, title, body string, labels []string) error {
	issueNum, err := c.findOpenIssue(ctx, title)
	if err != nil {
		return err
	}

	if issueNum > 0 {
		return c.updateIssue(ctx, issueNum, body)
	}
	return c.createIssue(ctx, title, body, labels)
}

// EnsureLabels creates any missing repository labels defined in the provided map.
// The map key is the label name; value is the hex color (without "#").
func (c *Client) EnsureLabels(ctx context.Context, requiredLabels map[string]string) error {
	existing, _, err := c.gh.Issues.ListLabels(ctx, c.cfg.Owner, c.cfg.Repo, nil)
	if err != nil {
		return fmt.Errorf("notify: list labels: %w", err)
	}

	existingSet := make(map[string]bool, len(existing))
	for _, l := range existing {
		existingSet[l.GetName()] = true
	}

	for name, color := range requiredLabels {
		if existingSet[name] {
			continue
		}
		label := &github.Label{Name: github.Ptr(name), Color: github.Ptr(color)}
		if _, _, err := c.gh.Issues.CreateLabel(ctx, c.cfg.Owner, c.cfg.Repo, label); err != nil {
			return fmt.Errorf("notify: create label %q: %w", name, err)
		}
	}
	return nil
}

// findOpenIssue returns the issue number of an open issue with the given title, or 0.
func (c *Client) findOpenIssue(ctx context.Context, title string) (int, error) {
	opts := &github.IssueListByRepoOptions{State: "open"}
	issues, _, err := c.gh.Issues.ListByRepo(ctx, c.cfg.Owner, c.cfg.Repo, opts)
	if err != nil {
		return 0, fmt.Errorf("notify: list issues: %w", err)
	}
	for _, issue := range issues {
		if issue.GetTitle() == title {
			return issue.GetNumber(), nil
		}
	}
	return 0, nil
}

func (c *Client) createIssue(ctx context.Context, title, body string, labels []string) error {
	req := &github.IssueRequest{Title: &title, Body: &body, Labels: &labels}
	_, _, err := c.gh.Issues.Create(ctx, c.cfg.Owner, c.cfg.Repo, req)
	if err != nil {
		return fmt.Errorf("notify: create issue: %w", err)
	}
	return nil
}

func (c *Client) updateIssue(ctx context.Context, number int, body string) error {
	req := &github.IssueRequest{Body: &body}
	_, _, err := c.gh.Issues.Edit(ctx, c.cfg.Owner, c.cfg.Repo, number, req)
	if err != nil {
		return fmt.Errorf("notify: update issue #%d: %w", number, err)
	}
	return nil
}

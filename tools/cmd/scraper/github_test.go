package main

import (
	"os"
	"testing"
)

// TestIsValidGitHubToken tests GitHub token validation
func TestIsValidGitHubToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "valid personal access token",
			token: "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456",
			want:  true,
		},
		{
			name:  "valid OAuth token",
			token: "gho_1234567890abcdefghijklmnopqrstuvwxyz123456",
			want:  true,
		},
		{
			name:  "valid user token",
			token: "ghu_1234567890abcdefghijklmnopqrstuvwxyz123456",
			want:  true,
		},
		{
			name:  "valid server token",
			token: "ghs_1234567890abcdefghijklmnopqrstuvwxyz123456",
			want:  true,
		},
		{
			name:  "valid refresh token",
			token: "ghr_1234567890abcdefghijklmnopqrstuvwxyz123456",
			want:  true,
		},
		{
			name:  "too short",
			token: "ghp_short",
			want:  false,
		},
		{
			name:  "invalid prefix",
			token: "xxx_1234567890abcdefghijklmnopqrstuvwxyz123456",
			want:  false,
		},
		{
			name:  "contains special characters",
			token: "ghp_1234567890abcdefgh$jklmnopqrstuvwxyz123456",
			want:  false,
		},
		{
			name:  "empty",
			token: "",
			want:  false,
		},
		{
			name:  "no prefix",
			token: "1234567890abcdefghijklmnopqrstuvwxyz123456789",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidGitHubToken(tt.token)
			if got != tt.want {
				t.Errorf("isValidGitHubToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLoadGitHubConfig_InvalidToken tests that invalid tokens are rejected
func TestLoadGitHubConfig_InvalidToken(t *testing.T) {
	// Save original env vars
	origToken := os.Getenv("GITHUB_TOKEN")
	origRepo := os.Getenv("GITHUB_REPOSITORY")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_REPOSITORY", origRepo)
	}()

	// Set invalid token
	os.Setenv("GITHUB_TOKEN", "invalid_token_format")
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")

	client, err := LoadGitHubConfig()
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
	if client != nil {
		t.Error("Expected nil client for invalid token")
	}
	if err != nil && err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

// TestLoadGitHubConfig_ValidToken tests that valid tokens are accepted
func TestLoadGitHubConfig_ValidToken(t *testing.T) {
	// Save original env vars
	origToken := os.Getenv("GITHUB_TOKEN")
	origRepo := os.Getenv("GITHUB_REPOSITORY")
	defer func() {
		os.Setenv("GITHUB_TOKEN", origToken)
		os.Setenv("GITHUB_REPOSITORY", origRepo)
	}()

	// Set valid token
	os.Setenv("GITHUB_TOKEN", "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456")
	os.Setenv("GITHUB_REPOSITORY", "owner/repo")

	client, err := LoadGitHubConfig()
	if err != nil {
		t.Errorf("Expected no error for valid token, got: %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil client for valid token")
	}
	if client != nil && client.Owner != "owner" {
		t.Errorf("Expected owner='owner', got '%s'", client.Owner)
	}
	if client != nil && client.Repo != "repo" {
		t.Errorf("Expected repo='repo', got '%s'", client.Repo)
	}
}

package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// httpClient is the HTTP client used for GitHub API calls.
// Package-level var for testability (swap in tests via t.Cleanup).
var httpClient = &http.Client{Timeout: 5 * time.Second}

// githubRelease represents the subset of GitHub's release response we need.
type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// fetchLatestRelease fetches the latest release from a GitHub repository.
// Supports optional GITHUB_TOKEN env var as Bearer token to avoid rate limits.
func fetchLatestRelease(ctx context.Context, owner, repo string) (githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return githubRelease{}, fmt.Errorf("build github request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return githubRelease{}, fmt.Errorf("github API request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// success — decode below
	case http.StatusForbidden:
		return githubRelease{}, fmt.Errorf("github API rate limit exceeded (HTTP 403)")
	case http.StatusNotFound:
		return githubRelease{}, fmt.Errorf("no releases found for %s/%s (HTTP 404)", owner, repo)
	default:
		return githubRelease{}, fmt.Errorf("github API returned HTTP %d for %s/%s", resp.StatusCode, owner, repo)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return githubRelease{}, fmt.Errorf("decode github release: %w", err)
	}

	return release, nil
}

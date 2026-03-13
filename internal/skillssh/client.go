package skillssh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// httpClient is the HTTP client used for skills.sh API calls.
// Package-level var for testability (swap in tests via t.Cleanup).
var httpClient = &http.Client{Timeout: 15 * time.Second}

// SearchSkill represents a single skill returned by the skills.sh search API.
type SearchSkill struct {
	ID       string `json:"id"`
	SkillID  string `json:"skillId"`
	Name     string `json:"name"`
	Installs int    `json:"installs"`
	Source   string `json:"source"`
}

type searchResponse struct {
	Skills []SearchSkill `json:"skills"`
}

// TrustedSources is the allowlist of skill sources used as a security proxy.
// Sources selected from the skills.sh leaderboard by install count and publisher reputation.
var TrustedSources = []string{
	// Tier 1 — official / platform publishers
	"vercel-labs/agent-skills",
	"vercel-labs/next-skills",
	"vercel-labs/agent-browser",
	"anthropics/skills",
	"remotion-dev/skills",
	"google-labs-code/stitch-skills",
	"microsoft/github-copilot-for-azure",
	"microsoft/azure-skills",
	// Tier 2 — well-known OSS publishers & products
	"github/awesome-copilot",
	"sleekdotdesign/agent-skills",
	"browser-use/browser-use",
	"obra/superpowers",
	"supabase/agent-skills",
	"squirrelscan/skills",
	"toolshell/skills",
	"better-auth/skills",
	"expo/skills",
	"wshobson/agents",
	"am-will/codex-skills",
	"callstackincubator/agent-skills",
	"antfu/skills",
	"roin-orca/skills",
}

// MinInstalls is the minimum install count a skill must have to be shown.
const MinInstalls = 500

// Search calls the skills.sh API and returns raw results (unfiltered).
// Returns empty slice (not error) on 404.
func Search(ctx context.Context, query string) ([]SearchSkill, error) {
	if query == "" {
		return []SearchSkill{}, nil
	}

	url := fmt.Sprintf("https://skills.sh/api/search?q=%s&limit=20", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build skills.sh request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("skills.sh API request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// success — decode below
	case http.StatusNotFound:
		return []SearchSkill{}, nil
	default:
		return nil, fmt.Errorf("skills.sh API returned HTTP %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode skills.sh response: %w", err)
	}

	return result.Skills, nil
}

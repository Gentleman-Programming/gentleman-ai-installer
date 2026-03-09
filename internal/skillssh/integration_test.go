package skillssh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// reactSkills are the trusted skills a React project should receive.
var reactSkills = []SearchSkill{
	{ID: "vercel-labs/agent-skills/react-best-practices", SkillID: "react-best-practices", Name: "react-best-practices", Installs: 12000, Source: "vercel-labs/agent-skills"},
	{ID: "vercel-labs/next-skills/nextjs-app-router", SkillID: "nextjs-app-router", Name: "nextjs-app-router", Installs: 8500, Source: "vercel-labs/next-skills"},
	{ID: "shady-org/bad-skills/react-exploit", SkillID: "react-exploit", Name: "react-exploit", Installs: 99999, Source: "shady-org/bad-skills"}, // untrusted — must be filtered
}

// goSkills are the trusted skills a Go project should receive.
var goSkills = []SearchSkill{
	{ID: "obra/superpowers/go-testing", SkillID: "go-testing", Name: "go-testing", Installs: 3200, Source: "obra/superpowers"},
	{ID: "antfu/skills/golang", SkillID: "golang", Name: "golang", Installs: 1800, Source: "antfu/skills"},
	{ID: "random-org/random/go-hacks", SkillID: "go-hacks", Name: "go-hacks", Installs: 50000, Source: "random-org/random"}, // untrusted — must be filtered
}

// newMockServer creates a test HTTP server that returns the provided skills for
// any request. It registers cleanup automatically.
func newMockServer(t *testing.T, skills []SearchSkill) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResponse{Skills: skills})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// pointClientAt replaces httpClient to route all requests to srv.
func pointClientAt(t *testing.T, srv *httptest.Server) {
	t.Helper()
	swapHTTPClient(t, &http.Client{
		Transport: rewriteTransport{base: srv.URL, inner: http.DefaultTransport},
		Timeout:   5 * time.Second,
	})
}

// TestIntegration_ReactProjectRecommendsTrustedReactSkills verifies the
// full Search+Filter pipeline for a React project:
//   - Only trusted sources pass through
//   - Results are sorted by installs descending
//   - The untrusted high-install skill is dropped
func TestIntegration_ReactProjectRecommendsTrustedReactSkills(t *testing.T) {
	srv := newMockServer(t, reactSkills)
	pointClientAt(t, srv)

	raw, err := Search(context.Background(), "react")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	filtered := Filter(raw)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 trusted React skills, got %d", len(filtered))
	}

	// Must be sorted by installs descending.
	if filtered[0].Installs < filtered[1].Installs {
		t.Errorf("skills not sorted: first=%d, second=%d", filtered[0].Installs, filtered[1].Installs)
	}

	// Untrusted source must not appear.
	for _, s := range filtered {
		if s.Source == "shady-org/bad-skills" {
			t.Errorf("untrusted source %q slipped through for React project", s.Source)
		}
	}

	// Both trusted skills must be present.
	ids := make(map[string]bool)
	for _, s := range filtered {
		ids[s.SkillID] = true
	}
	for _, want := range []string{"react-best-practices", "nextjs-app-router"} {
		if !ids[want] {
			t.Errorf("expected skill %q not found in filtered results", want)
		}
	}
}

// TestIntegration_GoProjectRecommendsTrustedGoSkills verifies the full
// Search+Filter pipeline for a Go project:
//   - Only trusted sources pass through
//   - Results are sorted by installs descending
//   - The untrusted high-install skill is dropped
func TestIntegration_GoProjectRecommendsTrustedGoSkills(t *testing.T) {
	srv := newMockServer(t, goSkills)
	pointClientAt(t, srv)

	raw, err := Search(context.Background(), "go")
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	filtered := Filter(raw)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 trusted Go skills, got %d", len(filtered))
	}

	if filtered[0].Installs < filtered[1].Installs {
		t.Errorf("skills not sorted: first=%d, second=%d", filtered[0].Installs, filtered[1].Installs)
	}

	for _, s := range filtered {
		if strings.Contains(s.Source, "random-org") {
			t.Errorf("untrusted source %q slipped through for Go project", s.Source)
		}
	}

	ids := make(map[string]bool)
	for _, s := range filtered {
		ids[s.SkillID] = true
	}
	for _, want := range []string{"go-testing", "golang"} {
		if !ids[want] {
			t.Errorf("expected skill %q not found in filtered results", want)
		}
	}
}

// TestIntegration_SearchErrorDoesNotPanic verifies that a non-200 response
// from the API returns an error without panicking.
func TestIntegration_SearchErrorDoesNotPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)
	pointClientAt(t, srv)

	_, err := Search(context.Background(), "react")
	if err == nil {
		t.Fatal("expected error on 503, got nil")
	}
}

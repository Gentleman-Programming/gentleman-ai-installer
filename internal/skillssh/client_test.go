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

// swapHTTPClient replaces the package-level httpClient for the duration of a
// test and restores it via t.Cleanup.
func swapHTTPClient(t *testing.T, client *http.Client) {
	t.Helper()
	orig := httpClient
	httpClient = client
	t.Cleanup(func() { httpClient = orig })
}

func TestSearch_Success(t *testing.T) {
	skills := []SearchSkill{
		{ID: "1", SkillID: "react-hooks", Name: "React Hooks", Installs: 500, Source: "vercel-labs/agent-skills"},
		{ID: "2", SkillID: "vue-basics", Name: "Vue Basics", Installs: 200, Source: "google-labs-code/stitch-skills"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResponse{Skills: skills})
	}))
	defer srv.Close()

	// Point the client at our test server by swapping the base URL via a
	// custom transport that rewrites the host.
	swapHTTPClient(t, srv.Client())

	// We need to override the URL too — simplest approach: use a custom
	// RoundTripper that rewrites requests to the test server.
	swapHTTPClient(t, &http.Client{
		Transport: rewriteTransport{base: srv.URL, inner: http.DefaultTransport},
		Timeout:   5 * time.Second,
	})

	got, err := Search(context.Background(), "react")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(got))
	}
	if got[0].SkillID != "react-hooks" {
		t.Errorf("first skill = %q, want %q", got[0].SkillID, "react-hooks")
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	// No HTTP call should be made for an empty query.
	got, err := Search(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d skills", len(got))
	}
}

func TestSearch_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	swapHTTPClient(t, &http.Client{
		Transport: rewriteTransport{base: srv.URL, inner: http.DefaultTransport},
		Timeout:   5 * time.Second,
	})

	got, err := Search(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("expected no error on 404, got: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice on 404, got %d", len(got))
	}
}

func TestSearch_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	swapHTTPClient(t, &http.Client{
		Transport: rewriteTransport{base: srv.URL, inner: http.DefaultTransport},
		Timeout:   5 * time.Second,
	})

	_, err := Search(context.Background(), "react")
	if err == nil {
		t.Fatal("expected error on HTTP 500, got nil")
	}
}

func TestSearch_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	swapHTTPClient(t, &http.Client{
		Transport: rewriteTransport{base: srv.URL, inner: http.DefaultTransport},
		Timeout:   5 * time.Second,
	})

	_, err := Search(context.Background(), "react")
	if err == nil {
		t.Fatal("expected error on malformed JSON, got nil")
	}
}

func TestSearch_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block long enough for the client timeout to fire.
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	swapHTTPClient(t, &http.Client{
		Transport: rewriteTransport{base: srv.URL, inner: http.DefaultTransport},
		Timeout:   50 * time.Millisecond,
	})

	_, err := Search(context.Background(), "react")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// rewriteTransport rewrites requests to always hit the given base URL,
// preserving path and query. This lets tests use httptest.NewServer with the
// package-level httpClient without DNS tricks.
type rewriteTransport struct {
	base  string
	inner http.RoundTripper
}

func (rt rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = "http"
	clone.URL.Host = strings.TrimPrefix(rt.base, "http://")
	return rt.inner.RoundTrip(clone)
}

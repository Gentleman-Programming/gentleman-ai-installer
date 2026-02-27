package engram

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyInstalled(t *testing.T) {
	original := lookPath
	t.Cleanup(func() { lookPath = original })

	lookPath = func(string) (string, error) { return "/opt/homebrew/bin/engram", nil }
	if err := VerifyInstalled(); err != nil {
		t.Fatalf("VerifyInstalled() error = %v", err)
	}

	lookPath = func(string) (string, error) { return "", errors.New("missing") }
	if err := VerifyInstalled(); err == nil {
		t.Fatalf("VerifyInstalled() expected missing binary error")
	}
}

func TestVerifyHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := VerifyHealth(context.Background(), server.URL); err != nil {
		t.Fatalf("VerifyHealth() error = %v", err)
	}

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer badServer.Close()

	if err := VerifyHealth(context.Background(), badServer.URL); err == nil {
		t.Fatalf("VerifyHealth() expected non-200 error")
	}
}

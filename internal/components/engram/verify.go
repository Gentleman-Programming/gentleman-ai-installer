package engram

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

var lookPath = exec.LookPath

func VerifyInstalled() error {
	if _, err := lookPath("engram"); err != nil {
		return fmt.Errorf("engram binary not found in PATH: %w", err)
	}

	return nil
}

func VerifyHealth(ctx context.Context, baseURL string) error {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://127.0.0.1:7437"
	}

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/health", nil)
	if err != nil {
		return fmt.Errorf("build engram health request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("engram health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("engram health check returned status %d", resp.StatusCode)
	}

	return nil
}

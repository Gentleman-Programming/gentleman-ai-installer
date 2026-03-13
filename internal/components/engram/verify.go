package engram

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	lookPath    = exec.LookPath
	execCommand = exec.Command
)

// InstallHint returns platform-specific installation instructions for Engram.
func InstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew tap Gentleman-Programming/homebrew-tap && brew install engram"
	case "windows":
		return "go install github.com/Gentleman-Programming/engram/cmd/engram@latest"
	default:
		// Linux and others
		return "go install github.com/Gentleman-Programming/engram/cmd/engram@latest"
	}
}

func VerifyInstalled() error {
	if _, err := lookPath("engram"); err != nil {
		return fmt.Errorf("engram binary not found in PATH.\n\nTo install engram:\n  %s\n\nNote: After installation, ensure engram is in your PATH (add $HOME/go/bin to your shell profile)", InstallHint())
	}

	return nil
}

// VerifyVersion runs "engram version" and returns the trimmed output.
// Returns an error if the command fails or produces no output.
func VerifyVersion() (string, error) {
	cmd := execCommand("engram", "version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("engram version command failed: %w", err)
	}

	version := strings.TrimSpace(string(out))
	if version == "" {
		return "", fmt.Errorf("engram version returned empty output")
	}

	return version, nil
}

// VerifyHealth checks if the Engram MCP server is responding.
// Returns an error if the server is not reachable.
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
		return fmt.Errorf("engram MCP server is not responding at %s.\n\nPossible causes:\n  - Engram is not running as an MCP server\n  - The MCP configuration is incorrect\n  - Port 7437 is not available\n\nTo start engram as MCP server:\n  engram mcp", baseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("engram health check returned status %d (expected 200)", resp.StatusCode)
	}

	return nil
}

// VerifyComplete runs all verification checks and returns a comprehensive result.
func VerifyComplete(ctx context.Context) error {
	// Check if engram binary exists
	if err := VerifyInstalled(); err != nil {
		return err
	}

	// Check if engram is running as MCP server
	if err := VerifyHealth(ctx, ""); err != nil {
		return fmt.Errorf("engram is installed but not accessible as MCP server.\n\n%v\n\nTo fix:\n  1. Start engram MCP server: engram mcp\n  2. Or restart your AI agent to reload MCP configurations", err)
	}

	return nil
}

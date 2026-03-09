package stackdetect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile is a test helper that writes content to a file inside dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
}

// containsAll verifies that all expected keywords appear in the result string.
func containsAll(t *testing.T, result string, expected []string) {
	t.Helper()
	for _, kw := range expected {
		if !strings.Contains(result, kw) {
			t.Errorf("result %q missing keyword %q", result, kw)
		}
	}
}

func TestDetect_PackageJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{
		"dependencies": {
			"react": "^18.0.0",
			"express": "^4.0.0"
		},
		"devDependencies": {
			"typescript": "^5.0.0",
			"vite": "^5.0.0"
		}
	}`)

	result := Detect(dir)

	containsAll(t, result, []string{"react", "express", "typescript", "vite"})
}

func TestDetect_PackageJSON_IgnoresUnknownPackages(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{
		"dependencies": {
			"lodash": "^4.0.0",
			"react": "^18.0.0"
		}
	}`)

	result := Detect(dir)

	if strings.Contains(result, "lodash") {
		t.Errorf("result %q should not contain unknown package 'lodash'", result)
	}
	if !strings.Contains(result, "react") {
		t.Errorf("result %q missing 'react'", result)
	}
}

func TestDetect_GoMod(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", `module github.com/example/myapp

go 1.22

require (
	github.com/charmbracelet/bubbletea v1.3.4
	github.com/gin-gonic/gin v1.9.0
)
`)

	result := Detect(dir)

	containsAll(t, result, []string{"bubbletea", "gin"})
}

func TestDetect_CargoToml(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cargo.toml", `[package]
name = "my-app"
version = "0.1.0"

[dependencies]
serde = { version = "1.0", features = ["derive"] }
tokio = "1.0"
axum = "0.7"
`)

	result := Detect(dir)

	containsAll(t, result, []string{"serde", "tokio", "axum"})
}

func TestDetect_PyprojectToml_Poetry(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", `[tool.poetry]
name = "my-project"

[tool.poetry.dependencies]
python = "^3.11"
fastapi = "^0.100.0"
pydantic = "^2.0"
`)

	result := Detect(dir)

	containsAll(t, result, []string{"fastapi", "pydantic"})
	if strings.Contains(result, "python") {
		t.Errorf("result %q should not contain 'python' (filtered out)", result)
	}
}

func TestDetect_RequirementsTxt(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "requirements.txt", `django>=4.0
requests==2.31.0
# this is a comment
celery~=5.3
flask[async]>=2.0
`)

	result := Detect(dir)

	containsAll(t, result, []string{"django", "requests", "celery", "flask"})
}

func TestDetect_ComposerJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "composer.json", `{
		"require": {
			"laravel/framework": "^10.0",
			"php": "^8.1",
			"symfony/console": "^6.0"
		}
	}`)

	result := Detect(dir)

	containsAll(t, result, []string{"laravel", "symfony"})
	if strings.Contains(result, "php") {
		t.Errorf("result %q should not contain 'php'", result)
	}
}

func TestDetect_MultiFileDedup(t *testing.T) {
	// Both go.mod and Cargo.toml include a package named "serde" (hypothetical)
	// More realistic: react appears in package.json deps and devDeps — should appear once.
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{
		"dependencies": { "react": "^18.0.0" },
		"devDependencies": { "react": "^18.0.0" }
	}`)

	result := Detect(dir)

	count := strings.Count(result, "react")
	if count != 1 {
		t.Errorf("'react' appears %d times in %q, want exactly 1", count, result)
	}
}

func TestDetect_NoFiles(t *testing.T) {
	dir := t.TempDir()

	result := Detect(dir)

	if result != "" {
		t.Errorf("expected empty string for empty dir, got %q", result)
	}
}

func TestDetect_MalformedJSONNoPanic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{not valid json at all`)
	writeFile(t, dir, "composer.json", `{also broken`)

	// Must not panic — just returns empty or partial results.
	result := Detect(dir)
	_ = result // result content doesn't matter, just must not panic
}

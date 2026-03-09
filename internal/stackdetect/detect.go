package stackdetect

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// knownJSFrameworks maps package.json dependency names to their keyword.
var knownJSFrameworks = map[string]string{
	"react":      "react",
	"vue":        "vue",
	"angular":    "angular",
	"@angular/core": "angular",
	"next":       "next",
	"next.js":    "next",
	"nuxt":       "nuxt",
	"svelte":     "svelte",
	"express":    "express",
	"fastify":    "fastify",
	"@nestjs/core": "nest",
	"nest":       "nest",
	"vite":       "vite",
	"webpack":    "webpack",
	"typescript": "typescript",
}

// Detect reads the given directory for well-known dependency files and extracts
// tech keywords. It never errors on missing or malformed files — those are
// silently skipped. Returns deduplicated, sorted, lowercase keywords as a
// space-joined string. Returns "" when no keywords are found.
func Detect(dir string) string {
	seen := make(map[string]struct{})

	addKeyword := func(kw string) {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw != "" {
			seen[kw] = struct{}{}
		}
	}

	parsePackageJSON(filepath.Join(dir, "package.json"), addKeyword)
	parseGoMod(filepath.Join(dir, "go.mod"), addKeyword)
	parseCargoToml(filepath.Join(dir, "Cargo.toml"), addKeyword)
	parsePyprojectToml(filepath.Join(dir, "pyproject.toml"), addKeyword)
	parseRequirementsTxt(filepath.Join(dir, "requirements.txt"), addKeyword)
	parseComposerJSON(filepath.Join(dir, "composer.json"), addKeyword)

	if len(seen) == 0 {
		return ""
	}

	keywords := make([]string, 0, len(seen))
	for kw := range seen {
		keywords = append(keywords, kw)
	}
	sort.Strings(keywords)

	return strings.Join(keywords, " ")
}

// parsePackageJSON extracts known framework names from dependencies and
// devDependencies.
func parsePackageJSON(path string, add func(string)) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var pkg struct {
		Dependencies    map[string]json.RawMessage `json:"dependencies"`
		DevDependencies map[string]json.RawMessage `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	for name := range pkg.Dependencies {
		if kw, ok := knownJSFrameworks[name]; ok {
			add(kw)
		}
	}
	for name := range pkg.DevDependencies {
		if kw, ok := knownJSFrameworks[name]; ok {
			add(kw)
		}
	}
}

// parseGoMod extracts the last two path segments from require directives as
// keywords (e.g. "github.com/gin-gonic/gin" → "gin").
func parseGoMod(path string, add func(string)) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inRequire := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}

		// single-line require: require github.com/foo/bar v1.2.3
		if strings.HasPrefix(line, "require ") {
			line = strings.TrimPrefix(line, "require ")
			line = strings.TrimSpace(line)
			addGoModEntry(line, add)
			continue
		}

		if inRequire {
			addGoModEntry(line, add)
		}
	}
}

// addGoModEntry extracts the last path segment from a go.mod require line.
func addGoModEntry(line string, add func(string)) {
	// strip inline comment
	if idx := strings.Index(line, "//"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	// line format: "module/path vX.Y.Z"
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}
	modPath := parts[0]
	segments := strings.Split(modPath, "/")
	if len(segments) == 0 {
		return
	}
	last := segments[len(segments)-1]
	if last != "" {
		add(last)
	}
}

// parseCargoToml extracts dependency names from the [dependencies] section of
// a Cargo.toml file (simple TOML parsing — not full spec).
func parseCargoToml(path string, add func(string)) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inDeps := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "[dependencies]" {
			inDeps = true
			continue
		}
		if inDeps && strings.HasPrefix(line, "[") {
			// entered a new section
			inDeps = false
			continue
		}
		if inDeps && line != "" && !strings.HasPrefix(line, "#") {
			// dependency line: name = "version" or name = { version = "..." }
			eq := strings.Index(line, "=")
			if eq > 0 {
				name := strings.TrimSpace(line[:eq])
				if name != "" {
					add(name)
				}
			}
		}
	}
}

// parsePyprojectToml extracts dependency names from [tool.poetry.dependencies]
// or [project.dependencies].
func parsePyprojectToml(path string, add func(string)) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inDeps := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "[tool.poetry.dependencies]" || line == "[project.dependencies]" {
			inDeps = true
			continue
		}
		if inDeps && strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}
		if inDeps && line != "" && !strings.HasPrefix(line, "#") {
			eq := strings.Index(line, "=")
			if eq > 0 {
				name := strings.TrimSpace(line[:eq])
				if name != "" && name != "python" {
					add(name)
				}
			}
		}
	}
}

// parseRequirementsTxt extracts package names from a requirements.txt file,
// stripping version specifiers (==, >=, <=, ~=, !=, >, <).
func parseRequirementsTxt(path string, add func(string)) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// strip inline comment
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// strip version specifiers — cut at first operator character
		for _, sep := range []string{"==", ">=", "<=", "~=", "!=", ">", "<", "["} {
			if idx := strings.Index(line, sep); idx >= 0 {
				line = strings.TrimSpace(line[:idx])
			}
		}

		if line != "" {
			add(line)
		}
	}
}

// parseComposerJSON extracts package keywords from the "require" key, stripping
// the vendor prefix (e.g. "laravel/framework" → "laravel").
func parseComposerJSON(path string, add func(string)) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var composer struct {
		Require map[string]json.RawMessage `json:"require"`
	}
	if err := json.Unmarshal(data, &composer); err != nil {
		return
	}

	for name := range composer.Require {
		// strip vendor prefix: "laravel/framework" → "laravel"
		// also ignore "php" and "ext-*" entries
		if name == "php" || strings.HasPrefix(name, "ext-") {
			continue
		}
		parts := strings.SplitN(name, "/", 2)
		keyword := parts[0]
		if keyword != "" {
			add(keyword)
		}
	}
}

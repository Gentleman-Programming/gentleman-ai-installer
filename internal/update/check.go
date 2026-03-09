package update

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/gentleman-programming/gentle-ai/internal/system"
)

// CheckAll runs update checks for all registered tools concurrently.
// currentVersion is the build-time version of gentle-ai (from app.Version).
// profile determines platform-specific update instructions.
func CheckAll(ctx context.Context, currentVersion string, profile system.PlatformProfile) []UpdateResult {
	results := make([]UpdateResult, len(Tools))

	var wg sync.WaitGroup
	for i, tool := range Tools {
		wg.Add(1)
		go func(idx int, t ToolInfo) {
			defer wg.Done()
			results[idx] = checkSingleTool(ctx, t, currentVersion, profile)
		}(i, tool)
	}

	wg.Wait()
	return results
}

// checkSingleTool checks a single tool: detects local version, fetches remote, compares.
func checkSingleTool(ctx context.Context, tool ToolInfo, currentBuildVersion string, profile system.PlatformProfile) UpdateResult {
	result := UpdateResult{Tool: tool}

	// Run local detection and remote fetch concurrently.
	var wg sync.WaitGroup
	var localVersion string
	var release githubRelease
	var fetchErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		localVersion = detectInstalledVersion(ctx, tool, currentBuildVersion)
	}()

	go func() {
		defer wg.Done()
		release, fetchErr = fetchLatestRelease(ctx, tool.Owner, tool.Repo)
	}()

	wg.Wait()

	result.InstalledVersion = localVersion
	result.UpdateHint = updateHint(tool, profile)

	// Handle fetch failure.
	if fetchErr != nil {
		result.Err = fetchErr
		if localVersion == "" && tool.DetectCmd != nil {
			result.Status = NotInstalled
		} else {
			result.Status = CheckFailed
		}
		return result
	}

	result.LatestVersion = normalizeVersion(release.TagName)
	result.ReleaseURL = release.HTMLURL

	// Determine status based on local version.
	if localVersion == "" {
		if tool.DetectCmd == nil {
			// gentle-ai with no build version (shouldn't happen, but handle gracefully).
			result.Status = VersionUnknown
		} else {
			// Binary not found on PATH.
			if _, err := lookPath(tool.DetectCmd[0]); err != nil {
				result.Status = NotInstalled
			} else {
				result.Status = VersionUnknown
			}
		}
		return result
	}

	// Check for non-semver local versions (e.g., "dev").
	normalizedLocal := normalizeVersion(localVersion)
	if !isSemver(normalizedLocal) {
		result.Status = VersionUnknown
		return result
	}

	// Compare versions.
	result.Status = compareVersions(normalizedLocal, result.LatestVersion)
	return result
}

// normalizeVersion strips a leading "v" and extracts a semver pattern.
func normalizeVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "v")

	match := versionRegexp.FindStringSubmatch(raw)
	if len(match) >= 2 {
		return match[1]
	}

	return raw
}

// isSemver checks if a version string looks like a semver (N.N or N.N.N).
func isSemver(v string) bool {
	return versionRegexp.MatchString(v)
}

// compareVersions returns UpToDate if local >= remote, UpdateAvailable otherwise.
func compareVersions(local, remote string) UpdateStatus {
	localParts := parseVersionParts(local)
	remoteParts := parseVersionParts(remote)

	for i := 0; i < 3; i++ {
		if localParts[i] > remoteParts[i] {
			return UpToDate
		}
		if localParts[i] < remoteParts[i] {
			return UpdateAvailable
		}
	}

	return UpToDate // equal
}

// parseVersionParts splits "1.2.3" into [1, 2, 3], padding with zeros.
// Same logic as internal/system/deps.go:parseVersionParts.
func parseVersionParts(version string) [3]int {
	parts := strings.SplitN(version, ".", 3)
	var result [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		n, _ := strconv.Atoi(parts[i])
		result[i] = n
	}
	return result
}

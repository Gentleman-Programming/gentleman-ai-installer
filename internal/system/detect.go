package system

import (
	"context"
	"os"
	"runtime"
	"strings"
)

type SystemInfo struct {
	OS        string
	Arch      string
	Shell     string
	Supported bool
	Profile   PlatformProfile
}

type PlatformProfile struct {
	OS             string
	LinuxDistro    string
	PackageManager string
	Supported      bool
}

const (
	LinuxDistroUnknown = "unknown"
	LinuxDistroUbuntu  = "ubuntu"
	LinuxDistroDebian  = "debian"
	LinuxDistroArch    = "arch"
)

type DetectionResult struct {
	System       SystemInfo
	Tools        map[string]ToolStatus
	Configs      []ConfigState
	Dependencies DependencyReport
}

func IsSupportedOS(goos string) bool {
	return goos == "darwin" || goos == "linux"
}

func Detect(ctx context.Context) (DetectionResult, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DetectionResult{}, err
	}

	tools := DetectTools(ctx, []string{"git", "curl", "brew", "node"})
	configs := ScanConfigs(homeDir)
	osReleaseContent, _ := osReleaseContent(runtime.GOOS)

	result := detectFromInputs(runtime.GOOS, runtime.GOARCH, os.Getenv("SHELL"), osReleaseContent, tools, configs)
	result.Dependencies = DetectDependencies(ctx, result.System.Profile)

	return result, nil
}

func detectFromInputs(goos, arch, shell, linuxOSRelease string, tools map[string]ToolStatus, configs []ConfigState) DetectionResult {
	if shell == "" {
		shell = "unknown"
	}

	profile := resolvePlatformProfile(goos, linuxOSRelease)

	return DetectionResult{
		System: SystemInfo{
			OS:        goos,
			Arch:      arch,
			Shell:     shell,
			Supported: profile.Supported,
			Profile:   profile,
		},
		Tools:   tools,
		Configs: configs,
	}
}

func osReleaseContent(goos string) (string, error) {
	if goos != "linux" {
		return "", nil
	}

	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func resolvePlatformProfile(goos, linuxOSRelease string) PlatformProfile {
	profile := PlatformProfile{OS: goos}

	switch goos {
	case "darwin":
		profile.PackageManager = "brew"
		profile.Supported = true
		return profile
	case "linux":
		distro := detectLinuxDistro(linuxOSRelease)
		profile.LinuxDistro = distro

		switch distro {
		case LinuxDistroUbuntu, LinuxDistroDebian:
			profile.PackageManager = "apt"
			profile.Supported = true
		case LinuxDistroArch:
			profile.PackageManager = "pacman"
			profile.Supported = true
		default:
			profile.PackageManager = ""
			profile.Supported = false
		}

		return profile
	default:
		profile.Supported = false
		return profile
	}
}

func detectLinuxDistro(linuxOSRelease string) string {
	if strings.TrimSpace(linuxOSRelease) == "" {
		return LinuxDistroUnknown
	}

	fields := map[string]string{}
	for _, line := range strings.Split(linuxOSRelease, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToUpper(strings.TrimSpace(parts[0]))
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		fields[key] = strings.ToLower(value)
	}

	id := fields["ID"]
	idLike := fields["ID_LIKE"]

	if isUbuntuLike(id, idLike) {
		if id == LinuxDistroDebian {
			return LinuxDistroDebian
		}
		return LinuxDistroUbuntu
	}

	if isArchLike(id, idLike) {
		return LinuxDistroArch
	}

	return LinuxDistroUnknown
}

func isUbuntuLike(id, idLike string) bool {
	if id == LinuxDistroUbuntu || id == LinuxDistroDebian {
		return true
	}

	for _, token := range strings.Fields(idLike) {
		if token == LinuxDistroUbuntu || token == LinuxDistroDebian {
			return true
		}
	}

	return false
}

func isArchLike(id, idLike string) bool {
	if id == LinuxDistroArch {
		return true
	}

	for _, token := range strings.Fields(idLike) {
		if token == LinuxDistroArch {
			return true
		}
	}

	return false
}

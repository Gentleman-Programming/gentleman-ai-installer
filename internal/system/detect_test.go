package system

import "testing"

func TestIsSupportedOS(t *testing.T) {
	tests := []struct {
		name string
		goos string
		want bool
	}{
		{name: "darwin is supported", goos: "darwin", want: true},
		{name: "linux is supported", goos: "linux", want: true},
		{name: "windows is unsupported", goos: "windows", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsSupportedOS(tc.goos)
			if got != tc.want {
				t.Fatalf("IsSupportedOS(%q) = %v, want %v", tc.goos, got, tc.want)
			}
		})
	}
}

func TestDetectFromInputsMarksSupportedMacOS(t *testing.T) {
	result := detectFromInputs("darwin", "arm64", "/bin/zsh", "", nil, nil)

	if !result.System.Supported {
		t.Fatalf("expected supported system for darwin")
	}

	if result.System.OS != "darwin" {
		t.Fatalf("expected OS darwin, got %q", result.System.OS)
	}

	if result.System.Profile.PackageManager != "brew" {
		t.Fatalf("expected brew package manager for macOS, got %q", result.System.Profile.PackageManager)
	}
}

func TestDetectFromInputsMarksUnsupportedLinuxDistro(t *testing.T) {
	osRelease := "ID=fedora\nID_LIKE=rhel fedora\n"
	result := detectFromInputs("linux", "amd64", "/bin/bash", osRelease, nil, nil)

	if result.System.Supported {
		t.Fatalf("expected unsupported system for unsupported linux distro")
	}

	if result.System.Profile.LinuxDistro != LinuxDistroUnknown {
		t.Fatalf("expected unknown distro, got %q", result.System.Profile.LinuxDistro)
	}
}

func TestDetectFromInputsMarksUbuntuSupported(t *testing.T) {
	osRelease := "ID=ubuntu\nID_LIKE=debian\n"
	result := detectFromInputs("linux", "amd64", "/bin/bash", osRelease, nil, nil)

	if !result.System.Supported {
		t.Fatalf("expected ubuntu linux to be supported")
	}

	if result.System.Profile.LinuxDistro != LinuxDistroUbuntu {
		t.Fatalf("expected ubuntu distro, got %q", result.System.Profile.LinuxDistro)
	}

	if result.System.Profile.PackageManager != "apt" {
		t.Fatalf("expected apt package manager, got %q", result.System.Profile.PackageManager)
	}
}

func TestDetectFromInputsMarksArchSupported(t *testing.T) {
	osRelease := "ID=arch\nID_LIKE=archlinux\n"
	result := detectFromInputs("linux", "amd64", "/bin/bash", osRelease, nil, nil)

	if !result.System.Supported {
		t.Fatalf("expected arch linux to be supported")
	}

	if result.System.Profile.LinuxDistro != LinuxDistroArch {
		t.Fatalf("expected arch distro, got %q", result.System.Profile.LinuxDistro)
	}

	if result.System.Profile.PackageManager != "pacman" {
		t.Fatalf("expected pacman package manager, got %q", result.System.Profile.PackageManager)
	}
}

// --- Batch E: Comprehensive platform detection matrix ---

func TestDetectLinuxDistroMatrix(t *testing.T) {
	tests := []struct {
		name       string
		osRelease  string
		wantDistro string
	}{
		{
			name:       "ubuntu 22.04",
			osRelease:  "ID=ubuntu\nID_LIKE=debian\nVERSION_ID=\"22.04\"\n",
			wantDistro: LinuxDistroUbuntu,
		},
		{
			name:       "debian 12",
			osRelease:  "ID=debian\nVERSION_ID=\"12\"\n",
			wantDistro: LinuxDistroDebian,
		},
		{
			name:       "linux mint derivative of ubuntu",
			osRelease:  "ID=linuxmint\nID_LIKE=\"ubuntu debian\"\nVERSION_ID=\"21.3\"\n",
			wantDistro: LinuxDistroUbuntu,
		},
		{
			name:       "pop os derivative of ubuntu",
			osRelease:  "ID=pop\nID_LIKE=\"ubuntu debian\"\n",
			wantDistro: LinuxDistroUbuntu,
		},
		{
			name:       "arch linux",
			osRelease:  "ID=arch\nID_LIKE=archlinux\n",
			wantDistro: LinuxDistroArch,
		},
		{
			name:       "manjaro derivative of arch",
			osRelease:  "ID=manjaro\nID_LIKE=arch\n",
			wantDistro: LinuxDistroArch,
		},
		{
			name:       "endeavouros derivative of arch",
			osRelease:  "ID=endeavouros\nID_LIKE=arch\n",
			wantDistro: LinuxDistroArch,
		},
		{
			name:       "fedora is unsupported",
			osRelease:  "ID=fedora\nID_LIKE=\"rhel fedora\"\n",
			wantDistro: LinuxDistroUnknown,
		},
		{
			name:       "centos is unsupported",
			osRelease:  "ID=centos\nID_LIKE=\"rhel fedora\"\n",
			wantDistro: LinuxDistroUnknown,
		},
		{
			name:       "empty os-release",
			osRelease:  "",
			wantDistro: LinuxDistroUnknown,
		},
		{
			name:       "whitespace-only os-release",
			osRelease:  "   \n  \n",
			wantDistro: LinuxDistroUnknown,
		},
		{
			name:       "comment-only os-release",
			osRelease:  "# This is a comment\n# Another comment\n",
			wantDistro: LinuxDistroUnknown,
		},
		{
			name:       "malformed lines are ignored",
			osRelease:  "no-equals-sign\nID=ubuntu\n",
			wantDistro: LinuxDistroUbuntu,
		},
		{
			name:       "quoted values are handled",
			osRelease:  "ID=\"ubuntu\"\nID_LIKE=\"debian\"\n",
			wantDistro: LinuxDistroUbuntu,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := detectLinuxDistro(tc.osRelease)
			if got != tc.wantDistro {
				t.Fatalf("detectLinuxDistro() = %q, want %q", got, tc.wantDistro)
			}
		})
	}
}

func TestResolvePlatformProfileMatrix(t *testing.T) {
	tests := []struct {
		name          string
		goos          string
		osRelease     string
		wantOS        string
		wantPM        string
		wantDistro    string
		wantSupported bool
	}{
		{
			name:          "darwin profile",
			goos:          "darwin",
			wantOS:        "darwin",
			wantPM:        "brew",
			wantSupported: true,
		},
		{
			name:          "ubuntu profile",
			goos:          "linux",
			osRelease:     "ID=ubuntu\nID_LIKE=debian\n",
			wantOS:        "linux",
			wantPM:        "apt",
			wantDistro:    LinuxDistroUbuntu,
			wantSupported: true,
		},
		{
			name:          "debian profile",
			goos:          "linux",
			osRelease:     "ID=debian\n",
			wantOS:        "linux",
			wantPM:        "apt",
			wantDistro:    LinuxDistroDebian,
			wantSupported: true,
		},
		{
			name:          "arch profile",
			goos:          "linux",
			osRelease:     "ID=arch\n",
			wantOS:        "linux",
			wantPM:        "pacman",
			wantDistro:    LinuxDistroArch,
			wantSupported: true,
		},
		{
			name:          "unsupported distro profile",
			goos:          "linux",
			osRelease:     "ID=fedora\n",
			wantOS:        "linux",
			wantPM:        "",
			wantDistro:    LinuxDistroUnknown,
			wantSupported: false,
		},
		{
			name:          "windows is unsupported",
			goos:          "windows",
			wantOS:        "windows",
			wantPM:        "",
			wantSupported: false,
		},
		{
			name:          "linux without os-release is unsupported",
			goos:          "linux",
			osRelease:     "",
			wantOS:        "linux",
			wantPM:        "",
			wantDistro:    LinuxDistroUnknown,
			wantSupported: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			profile := resolvePlatformProfile(tc.goos, tc.osRelease)
			if profile.OS != tc.wantOS {
				t.Fatalf("OS = %q, want %q", profile.OS, tc.wantOS)
			}
			if profile.PackageManager != tc.wantPM {
				t.Fatalf("PackageManager = %q, want %q", profile.PackageManager, tc.wantPM)
			}
			if profile.LinuxDistro != tc.wantDistro {
				t.Fatalf("LinuxDistro = %q, want %q", profile.LinuxDistro, tc.wantDistro)
			}
			if profile.Supported != tc.wantSupported {
				t.Fatalf("Supported = %v, want %v", profile.Supported, tc.wantSupported)
			}
		})
	}
}

func TestDetectFromInputsShellDefaultsToUnknown(t *testing.T) {
	result := detectFromInputs("darwin", "arm64", "", "", nil, nil)
	if result.System.Shell != "unknown" {
		t.Fatalf("Shell = %q, want %q", result.System.Shell, "unknown")
	}
}

func TestDetectFromInputsProfileIsPopulatedInSystem(t *testing.T) {
	osRelease := "ID=ubuntu\nID_LIKE=debian\n"
	result := detectFromInputs("linux", "amd64", "/bin/bash", osRelease, nil, nil)

	if result.System.Profile.OS != "linux" {
		t.Fatalf("Profile.OS = %q, want linux", result.System.Profile.OS)
	}
	if result.System.Profile.LinuxDistro != LinuxDistroUbuntu {
		t.Fatalf("Profile.LinuxDistro = %q, want ubuntu", result.System.Profile.LinuxDistro)
	}
	if result.System.Profile.PackageManager != "apt" {
		t.Fatalf("Profile.PackageManager = %q, want apt", result.System.Profile.PackageManager)
	}
	if !result.System.Profile.Supported {
		t.Fatalf("Profile.Supported = false, want true")
	}
	// System.Supported should mirror profile
	if result.System.Supported != result.System.Profile.Supported {
		t.Fatalf("System.Supported (%v) != Profile.Supported (%v)", result.System.Supported, result.System.Profile.Supported)
	}
}

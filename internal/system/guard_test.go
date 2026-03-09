package system

import (
	"errors"
	"strings"
	"testing"
)

func TestEnsureSupportedOSAllowsMacOS(t *testing.T) {
	if err := EnsureSupportedOS("darwin"); err != nil {
		t.Fatalf("expected no error for macOS, got %v", err)
	}
}

func TestEnsureSupportedOSAllowsWindows(t *testing.T) {
	if err := EnsureSupportedOS("windows"); err != nil {
		t.Fatalf("expected no error for Windows, got %v", err)
	}
}

func TestEnsureSupportedOSRejectsUnsupported(t *testing.T) {
	err := EnsureSupportedOS("freebsd")
	if err == nil {
		t.Fatalf("expected error for unsupported OS")
	}

	if !errors.Is(err, ErrUnsupportedOS) {
		t.Fatalf("expected ErrUnsupportedOS, got %v", err)
	}

	if !strings.Contains(err.Error(), "only macOS, Linux, and Windows are supported") {
		t.Fatalf("expected explicit OS support message, got %q", err.Error())
	}
}

func TestEnsureSupportedPlatformAllowsSupportedLinux(t *testing.T) {
	err := EnsureSupportedPlatform(PlatformProfile{OS: "linux", LinuxDistro: LinuxDistroUbuntu, Supported: true})
	if err != nil {
		t.Fatalf("expected ubuntu profile to be supported, got %v", err)
	}
}

func TestEnsureSupportedPlatformRejectsUnsupportedLinuxDistro(t *testing.T) {
	err := EnsureSupportedPlatform(PlatformProfile{OS: "linux", LinuxDistro: "gentoo", Supported: false})
	if err == nil {
		t.Fatalf("expected error for unsupported linux distro")
	}

	if !errors.Is(err, ErrUnsupportedLinuxDistro) {
		t.Fatalf("expected ErrUnsupportedLinuxDistro, got %v", err)
	}

	if !strings.Contains(err.Error(), "Linux support is limited to Ubuntu/Debian, Arch, and Fedora/RHEL") {
		t.Fatalf("expected distro guard message, got %q", err.Error())
	}
}

func TestEnsureSupportedPlatformAllowsFedora(t *testing.T) {
	err := EnsureSupportedPlatform(PlatformProfile{OS: "linux", LinuxDistro: LinuxDistroFedora, Supported: true})
	if err != nil {
		t.Fatalf("expected fedora profile to be supported, got %v", err)
	}
}

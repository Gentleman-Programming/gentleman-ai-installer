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

func TestEnsureSupportedOSRejectsNonMacOS(t *testing.T) {
	err := EnsureSupportedOS("windows")
	if err == nil {
		t.Fatalf("expected error for unsupported OS")
	}

	if !errors.Is(err, ErrUnsupportedOS) {
		t.Fatalf("expected ErrUnsupportedOS, got %v", err)
	}

	if !strings.Contains(err.Error(), "only macOS and Linux are supported in v0.2.0") {
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
	err := EnsureSupportedPlatform(PlatformProfile{OS: "linux", LinuxDistro: "fedora", Supported: false})
	if err == nil {
		t.Fatalf("expected error for unsupported linux distro")
	}

	if !errors.Is(err, ErrUnsupportedLinuxDistro) {
		t.Fatalf("expected ErrUnsupportedLinuxDistro, got %v", err)
	}

	if !strings.Contains(err.Error(), "Linux support is limited to Ubuntu/Debian and Arch") {
		t.Fatalf("expected distro guard message, got %q", err.Error())
	}
}

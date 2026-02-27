package system

import (
	"errors"
	"fmt"
	"runtime"
)

var ErrUnsupportedOS = errors.New("unsupported operating system")
var ErrUnsupportedLinuxDistro = errors.New("unsupported linux distro")

func EnsureCurrentOSSupported() error {
	return EnsureSupportedOS(runtime.GOOS)
}

func EnsureSupportedOS(goos string) error {
	if IsSupportedOS(goos) {
		return nil
	}

	return fmt.Errorf("%w: only macOS and Linux are supported in v0.2.0 (detected %s)", ErrUnsupportedOS, goos)
}

func EnsureSupportedPlatform(profile PlatformProfile) error {
	if err := EnsureSupportedOS(profile.OS); err != nil {
		return err
	}

	if profile.OS == "linux" && !profile.Supported {
		return fmt.Errorf("%w: Linux support is limited to Ubuntu/Debian and Arch (detected %s)", ErrUnsupportedLinuxDistro, profile.LinuxDistro)
	}

	return nil
}

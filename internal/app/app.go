package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/gentleman-programming/gentleman-ai-installer/internal/cli"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/system"
	"github.com/gentleman-programming/gentleman-ai-installer/internal/verify"
)

func Run() error {
	return RunArgs(os.Args[1:], os.Stdout)
}

func RunArgs(args []string, stdout io.Writer) error {
	if err := system.EnsureCurrentOSSupported(); err != nil {
		return err
	}

	result, err := system.Detect(context.Background())
	if err != nil {
		return fmt.Errorf("detect system: %w", err)
	}

	if !result.System.Supported {
		return system.EnsureSupportedPlatform(result.System.Profile)
	}

	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "install":
		installResult, err := cli.RunInstall(args[1:], result)
		if err != nil {
			return err
		}

		if installResult.DryRun {
			_, _ = fmt.Fprintln(stdout, cli.RenderDryRun(installResult))
		} else {
			_, _ = fmt.Fprint(stdout, verify.RenderReport(installResult.Verify))
		}

		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

package app

import (
	"context"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/internal/cli"
	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/tui"
	"github.com/gentleman-programming/gentle-ai/internal/verify"
)

// Version is set from main via ldflags at build time.
var Version = "dev"

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
		m := tui.NewModel(result, Version)
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err := p.Run()
		return err
	}

	switch args[0] {
	case "version", "--version", "-v":
		_, _ = fmt.Fprintf(stdout, "gentle-ai %s\n", Version)
		return nil
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

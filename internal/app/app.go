package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/internal/backup"
	"github.com/gentleman-programming/gentle-ai/internal/cli"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/pipeline"
	"github.com/gentleman-programming/gentle-ai/internal/planner"
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
		m.ExecuteFn = tuiExecute
		m.RestoreFn = tuiRestore
		m.Backups = ListBackups()
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

// tuiExecute creates a real install runtime and runs the pipeline with progress reporting.
func tuiExecute(
	selection model.Selection,
	resolved planner.ResolvedPlan,
	detection system.DetectionResult,
	onProgress pipeline.ProgressFunc,
) pipeline.ExecutionResult {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return pipeline.ExecutionResult{Err: fmt.Errorf("resolve user home directory: %w", err)}
	}

	profile := cli.ResolveInstallProfile(detection)
	resolved.PlatformDecision = planner.PlatformDecisionFromProfile(profile)

	stagePlan, err := cli.BuildRealStagePlan(homeDir, selection, resolved, profile)
	if err != nil {
		return pipeline.ExecutionResult{Err: fmt.Errorf("build stage plan: %w", err)}
	}

	orchestrator := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithFailurePolicy(pipeline.ContinueOnError),
		pipeline.WithProgressFunc(onProgress),
	)

	return orchestrator.Execute(stagePlan)
}

// tuiRestore restores a backup from its manifest.
func tuiRestore(manifest backup.Manifest) error {
	return backup.RestoreService{}.Restore(manifest)
}

// ListBackups returns all backup manifests from the backup directory.
func ListBackups() []backup.Manifest {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	backupRoot := filepath.Join(homeDir, ".gentle-ai", "backups")
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return nil
	}

	manifests := make([]backup.Manifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(backupRoot, entry.Name(), backup.ManifestFilename)
		manifest, err := backup.ReadManifest(manifestPath)
		if err != nil {
			continue
		}
		manifests = append(manifests, manifest)
	}

	// Sort by creation time (newest first) — the IDs are timestamps.
	for i := 0; i < len(manifests); i++ {
		for j := i + 1; j < len(manifests); j++ {
			if manifests[j].CreatedAt.After(manifests[i].CreatedAt) {
				manifests[i], manifests[j] = manifests[j], manifests[i]
			}
		}
	}

	return manifests
}

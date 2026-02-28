package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/agents"
	"github.com/gentleman-programming/gentle-ai/internal/backup"
	"github.com/gentleman-programming/gentle-ai/internal/components/engram"
	"github.com/gentleman-programming/gentle-ai/internal/components/gga"
	"github.com/gentleman-programming/gentle-ai/internal/components/mcp"
	"github.com/gentleman-programming/gentle-ai/internal/components/permissions"
	"github.com/gentleman-programming/gentle-ai/internal/components/persona"
	"github.com/gentleman-programming/gentle-ai/internal/components/sdd"
	"github.com/gentleman-programming/gentle-ai/internal/components/skills"
	"github.com/gentleman-programming/gentle-ai/internal/components/theme"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/pipeline"
	"github.com/gentleman-programming/gentle-ai/internal/planner"
	"github.com/gentleman-programming/gentle-ai/internal/system"
	"github.com/gentleman-programming/gentle-ai/internal/verify"
)

type InstallResult struct {
	Selection model.Selection
	Resolved  planner.ResolvedPlan
	Review    planner.ReviewPayload
	Plan      pipeline.StagePlan
	Execution pipeline.ExecutionResult
	Verify    verify.Report
	DryRun    bool
}

var (
	osUserHomeDir = os.UserHomeDir
	runCommand    = executeCommand
)

func RunInstall(args []string, detection system.DetectionResult) (InstallResult, error) {
	flags, err := ParseInstallFlags(args)
	if err != nil {
		return InstallResult{}, err
	}

	input, err := NormalizeInstallFlags(flags, detection)
	if err != nil {
		return InstallResult{}, err
	}

	resolved, err := planner.NewResolver(planner.MVPGraph()).Resolve(input.Selection)
	if err != nil {
		return InstallResult{}, err
	}
	profile := ResolveInstallProfile(detection)
	resolved.PlatformDecision = planner.PlatformDecisionFromProfile(profile)

	review := planner.BuildReviewPayload(input.Selection, resolved)
	stagePlan := buildStagePlan(input.Selection, resolved)

	result := InstallResult{
		Selection: input.Selection,
		Resolved:  resolved,
		Review:    review,
		Plan:      stagePlan,
		DryRun:    input.DryRun,
	}

	if input.DryRun {
		return result, nil
	}

	homeDir, err := osUserHomeDir()
	if err != nil {
		return result, fmt.Errorf("resolve user home directory: %w", err)
	}

	runtime, err := newInstallRuntime(homeDir, input.Selection, resolved, profile)
	if err != nil {
		return result, err
	}

	stagePlan = runtime.stagePlan()
	result.Plan = stagePlan

	orchestrator := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result.Execution = orchestrator.Execute(stagePlan)
	if result.Execution.Err != nil {
		return result, fmt.Errorf("execute install pipeline: %w", result.Execution.Err)
	}

	result.Verify = runPostApplyVerification(homeDir, input.Selection, resolved)
	if !result.Verify.Ready {
		return result, fmt.Errorf("post-apply verification failed:\n%s", verify.RenderReport(result.Verify))
	}

	return result, nil
}

func buildStagePlan(selection model.Selection, resolved planner.ResolvedPlan) pipeline.StagePlan {
	prepare := []pipeline.Step{noopStep{id: "prepare:system-check"}}
	apply := make([]pipeline.Step, 0, len(resolved.Agents)+len(resolved.OrderedComponents))

	for _, agent := range resolved.Agents {
		apply = append(apply, noopStep{id: "agent:" + string(agent)})
	}

	for _, component := range resolved.OrderedComponents {
		apply = append(apply, noopStep{id: "component:" + string(component)})
	}

	if len(selection.Agents) == 0 && len(resolved.OrderedComponents) == 0 {
		prepare = nil
	}

	return pipeline.StagePlan{Prepare: prepare, Apply: apply}
}

type installRuntime struct {
	homeDir    string
	selection  model.Selection
	resolved   planner.ResolvedPlan
	profile    system.PlatformProfile
	backupRoot string
	state      *runtimeState
}

type runtimeState struct {
	manifest backup.Manifest
}

func newInstallRuntime(homeDir string, selection model.Selection, resolved planner.ResolvedPlan, profile system.PlatformProfile) (*installRuntime, error) {
	backupRoot := filepath.Join(homeDir, ".gentle-ai", "backups")
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create backup root directory %q: %w", backupRoot, err)
	}

	return &installRuntime{
		homeDir:    homeDir,
		selection:  selection,
		resolved:   resolved,
		profile:    profile,
		backupRoot: backupRoot,
		state:      &runtimeState{},
	}, nil
}

func (r *installRuntime) stagePlan() pipeline.StagePlan {
	targets := backupTargets(r.homeDir, r.selection, r.resolved)
	prepare := []pipeline.Step{
		prepareBackupStep{
			id:          "prepare:backup-snapshot",
			snapshotter: backup.NewSnapshotter(),
			snapshotDir: filepath.Join(r.backupRoot, time.Now().UTC().Format("20060102150405.000000000")),
			targets:     targets,
			state:       r.state,
		},
	}

	apply := make([]pipeline.Step, 0, len(r.resolved.Agents)+len(r.resolved.OrderedComponents)+1)
	apply = append(apply, rollbackRestoreStep{id: "apply:rollback-restore", state: r.state})

	for _, agent := range r.resolved.Agents {
		apply = append(apply, agentInstallStep{id: "agent:" + string(agent), agent: agent, profile: r.profile})
	}

	for _, component := range r.resolved.OrderedComponents {
		apply = append(apply, componentApplyStep{
			id:        "component:" + string(component),
			component: component,
			homeDir:   r.homeDir,
			agents:    r.resolved.Agents,
			selection: r.selection,
			profile:   r.profile,
		})
	}

	return pipeline.StagePlan{Prepare: prepare, Apply: apply}
}

type prepareBackupStep struct {
	id          string
	snapshotter backup.Snapshotter
	snapshotDir string
	targets     []string
	state       *runtimeState
}

func (s prepareBackupStep) ID() string {
	return s.id
}

func (s prepareBackupStep) Run() error {
	manifest, err := s.snapshotter.Create(s.snapshotDir, s.targets)
	if err != nil {
		return fmt.Errorf("create backup snapshot: %w", err)
	}

	s.state.manifest = manifest
	return nil
}

type rollbackRestoreStep struct {
	id    string
	state *runtimeState
}

func (s rollbackRestoreStep) ID() string {
	return s.id
}

func (s rollbackRestoreStep) Run() error {
	return nil
}

func (s rollbackRestoreStep) Rollback() error {
	if len(s.state.manifest.Entries) == 0 {
		return nil
	}

	return backup.RestoreService{}.Restore(s.state.manifest)
}

type agentInstallStep struct {
	id      string
	agent   model.AgentID
	profile system.PlatformProfile
}

func (s agentInstallStep) ID() string {
	return s.id
}

func (s agentInstallStep) Run() error {
	adapter, err := agents.NewAdapter(s.agent)
	if err != nil {
		return fmt.Errorf("create adapter for %q: %w", s.agent, err)
	}

	commands, err := adapter.InstallCommand(s.profile)
	if err != nil {
		return fmt.Errorf("resolve install command for %q: %w", s.agent, err)
	}

	return runCommandSequence(commands)
}

type componentApplyStep struct {
	id        string
	component model.ComponentID
	homeDir   string
	agents    []model.AgentID
	selection model.Selection
	profile   system.PlatformProfile
}

func (s componentApplyStep) ID() string {
	return s.id
}

func (s componentApplyStep) Run() error {
	switch s.component {
	case model.ComponentEngram:
		commands, err := engram.InstallCommand(s.profile)
		if err != nil {
			return fmt.Errorf("resolve install command for component %q: %w", s.component, err)
		}
		if err := runCommandSequence(commands); err != nil {
			return err
		}
		for _, agent := range s.agents {
			if _, err := engram.Inject(s.homeDir, agent); err != nil {
				return fmt.Errorf("inject engram for %q: %w", agent, err)
			}
		}
		return nil
	case model.ComponentContext7:
		for _, agent := range s.agents {
			if _, err := mcp.Inject(s.homeDir, agent); err != nil {
				return fmt.Errorf("inject context7 for %q: %w", agent, err)
			}
		}
		return nil
	case model.ComponentPersona:
		for _, agent := range s.agents {
			if _, err := persona.Inject(s.homeDir, agent, s.selection.Persona); err != nil {
				return fmt.Errorf("inject persona for %q: %w", agent, err)
			}
		}
		return nil
	case model.ComponentPermission:
		for _, agent := range s.agents {
			if _, err := permissions.Inject(s.homeDir, agent); err != nil {
				return fmt.Errorf("inject permissions for %q: %w", agent, err)
			}
		}
		return nil
	case model.ComponentSDD:
		for _, agent := range s.agents {
			if _, err := sdd.Inject(s.homeDir, agent); err != nil {
				return fmt.Errorf("inject sdd for %q: %w", agent, err)
			}
		}
		return nil
	case model.ComponentSkills:
		skillIDs := selectedSkillIDs(s.selection)
		if len(skillIDs) == 0 {
			return nil
		}
		for _, agent := range s.agents {
			if _, err := skills.Inject(s.homeDir, agent, skillIDs); err != nil {
				return fmt.Errorf("inject skills for %q: %w", agent, err)
			}
		}
		return nil
	case model.ComponentGGA:
		commands, err := gga.InstallCommand(s.profile)
		if err != nil {
			return fmt.Errorf("resolve install command for component %q: %w", s.component, err)
		}
		if err := runCommandSequence(commands); err != nil {
			return err
		}
		if _, err := gga.WriteDefaultConfig(s.homeDir); err != nil {
			return fmt.Errorf("write gga config: %w", err)
		}
		return nil
	case model.ComponentTheme:
		for _, agent := range s.agents {
			if _, err := theme.Inject(s.homeDir, agent); err != nil {
				return fmt.Errorf("inject theme for %q: %w", agent, err)
			}
		}
		return nil
	default:
		return fmt.Errorf("component %q is not supported in install runtime", s.component)
	}
}

// BuildRealStagePlan creates a StagePlan with real backup, agent install, and component apply steps.
// It is used by both the CLI and TUI paths.
func BuildRealStagePlan(homeDir string, selection model.Selection, resolved planner.ResolvedPlan, profile system.PlatformProfile) (pipeline.StagePlan, error) {
	backupRoot := filepath.Join(homeDir, ".gentle-ai", "backups")
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return pipeline.StagePlan{}, fmt.Errorf("create backup root directory %q: %w", backupRoot, err)
	}

	runtime, err := newInstallRuntime(homeDir, selection, resolved, profile)
	if err != nil {
		return pipeline.StagePlan{}, err
	}

	return runtime.stagePlan(), nil
}

// ResolveInstallProfile returns the platform profile from detection, defaulting to darwin/brew.
func ResolveInstallProfile(detection system.DetectionResult) system.PlatformProfile {
	if detection.System.Profile.OS != "" {
		return detection.System.Profile
	}

	return system.PlatformProfile{
		OS:             "darwin",
		PackageManager: "brew",
		Supported:      true,
	}
}

// runCommandSequence runs each command in the sequence one at a time, stopping on first error.
func runCommandSequence(commands [][]string) error {
	if len(commands) == 0 {
		return fmt.Errorf("empty command sequence")
	}

	for _, command := range commands {
		if len(command) == 0 {
			return fmt.Errorf("empty command in sequence")
		}

		if err := runCommand(command[0], command[1:]...); err != nil {
			return fmt.Errorf("run command %q: %w", strings.Join(command, " "), err)
		}
	}

	return nil
}

func executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// selectedSkillIDs returns the skill IDs to install. If the selection
// has explicit skills, those are used; otherwise skills are derived from the preset.
func selectedSkillIDs(selection model.Selection) []model.SkillID {
	if len(selection.Skills) > 0 {
		return selection.Skills
	}

	return skills.SkillsForPreset(selection.Preset)
}

func backupTargets(homeDir string, selection model.Selection, resolved planner.ResolvedPlan) []string {
	paths := map[string]struct{}{}

	for _, component := range resolved.OrderedComponents {
		for _, path := range componentPaths(homeDir, selection, resolved.Agents, component) {
			paths[path] = struct{}{}
		}
	}

	targets := make([]string, 0, len(paths))
	for path := range paths {
		targets = append(targets, path)
	}

	return targets
}

func componentPaths(homeDir string, selection model.Selection, agents []model.AgentID, component model.ComponentID) []string {
	paths := []string{}
	for _, agent := range agents {
		switch component {
		case model.ComponentEngram:
			if agent == model.AgentClaudeCode {
				paths = append(paths,
					filepath.Join(homeDir, ".claude", "mcp", "engram.json"),
					filepath.Join(homeDir, ".claude", "CLAUDE.md"),
				)
			}
			if agent == model.AgentOpenCode {
				paths = append(paths, filepath.Join(homeDir, ".config", "opencode", "settings.json"))
			}
		case model.ComponentSDD:
			if agent == model.AgentClaudeCode {
				paths = append(paths, filepath.Join(homeDir, ".claude", "CLAUDE.md"))
			}
			if agent == model.AgentOpenCode {
				for _, command := range sdd.OpenCodeCommands() {
					paths = append(paths, filepath.Join(homeDir, ".config", "opencode", "commands", command.Name+".md"))
				}
			}
		case model.ComponentSkills:
			for _, skillID := range selectedSkillIDs(selection) {
				path, pathErr := skills.SkillPathForAgent(homeDir, agent, skillID)
				if pathErr == nil {
					paths = append(paths, path)
				}
			}
		case model.ComponentContext7:
			if agent == model.AgentClaudeCode {
				paths = append(paths, filepath.Join(homeDir, ".claude", "mcp", "context7.json"))
			}
			if agent == model.AgentOpenCode {
				paths = append(paths, filepath.Join(homeDir, ".config", "opencode", "settings.json"))
			}
		case model.ComponentPersona:
			// Custom persona does nothing — no files to backup or verify.
			if selection.Persona == model.PersonaCustom {
				break
			}
			if agent == model.AgentClaudeCode {
				paths = append(paths, filepath.Join(homeDir, ".claude", "CLAUDE.md"))
				if selection.Persona == model.PersonaGentleman {
					paths = append(paths,
						filepath.Join(homeDir, ".claude", "output-styles", "gentleman.md"),
						filepath.Join(homeDir, ".claude", "settings.json"),
					)
				}
			}
			if agent == model.AgentOpenCode {
				paths = append(paths, filepath.Join(homeDir, ".config", "opencode", "AGENTS.md"))
			}
		case model.ComponentPermission:
			if agent == model.AgentClaudeCode {
				paths = append(paths, filepath.Join(homeDir, ".claude", "settings.json"))
			}
			if agent == model.AgentOpenCode {
				paths = append(paths, filepath.Join(homeDir, ".config", "opencode", "settings.json"))
			}
		case model.ComponentGGA:
			paths = append(paths, filepath.Join(homeDir, ".config", "gga", "config.json"))
		case model.ComponentTheme:
			if agent == model.AgentClaudeCode {
				paths = append(paths, filepath.Join(homeDir, ".claude", "settings.json"))
			}
			if agent == model.AgentOpenCode {
				paths = append(paths, filepath.Join(homeDir, ".config", "opencode", "settings.json"))
			}
		}
	}

	return paths
}

func runPostApplyVerification(homeDir string, selection model.Selection, resolved planner.ResolvedPlan) verify.Report {
	checks := make([]verify.Check, 0)

	for _, component := range resolved.OrderedComponents {
		for _, path := range componentPaths(homeDir, selection, resolved.Agents, component) {
			currentPath := path
			checks = append(checks, verify.Check{
				ID:          "verify:file:" + currentPath,
				Description: "required file exists",
				Run: func(context.Context) error {
					if _, err := os.Stat(currentPath); err != nil {
						return err
					}
					return nil
				},
			})
		}
	}

	return verify.BuildReport(verify.RunChecks(context.Background(), checks))
}

type noopStep struct {
	id string
}

func (s noopStep) ID() string {
	return s.id
}

func (s noopStep) Run() error {
	return nil
}

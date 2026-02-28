#!/usr/bin/env bash
# e2e_test.sh — End-to-end tests for gentleman-ai installer
#
# Test tiers (controlled by environment variables):
#   (default)            Tier 1: binary existence + dry-run tests (fast, no side-effects)
#   RUN_FULL_E2E=1       Tier 2: full install tests (writes to filesystem)
#   RUN_BACKUP_TESTS=1   Tier 3: backup/restore tests
#
# Usage inside Docker:
#   ./e2e_test.sh                         # Tier 1 only
#   RUN_FULL_E2E=1 ./e2e_test.sh          # Tier 1 + 2
#   RUN_BACKUP_TESTS=1 ./e2e_test.sh      # Tier 1 + 3
#   RUN_FULL_E2E=1 RUN_BACKUP_TESTS=1 ./e2e_test.sh  # All tiers
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

# ---------------------------------------------------------------------------
# Resolve binary
# ---------------------------------------------------------------------------
BINARY="$(resolve_binary)"
if [ -z "$BINARY" ]; then
    echo "ERROR: gentleman-ai binary not found. Build it first."
    exit 1
fi
log_info "Using binary: $BINARY"

# ===========================================================================
# TIER 1 — Basic binary & dry-run tests (always run)
# ===========================================================================

test_binary_exists() {
    log_test "Binary exists and is executable"

    if [ -x "$(command -v "$BINARY")" ] || [ -x "$BINARY" ]; then
        log_pass "Binary is executable"
    else
        log_fail "Binary not found or not executable"
    fi
}

test_binary_runs() {
    log_test "Binary runs without panic"

    # The binary with no args enters TUI mode. Use 'install --dry-run' to get
    # non-interactive output.
    if output=$($BINARY install --dry-run 2>&1); then
        log_pass "Binary exited cleanly with --dry-run"
    else
        # Exit code != 0 is acceptable if it's a known validation error, but
        # a segfault/panic is not.
        if echo "$output" | grep -qi "panic"; then
            log_fail "Binary panicked: $output"
        else
            log_pass "Binary exited with non-zero (no panic)"
        fi
    fi
}

test_dry_run_output_format() {
    log_test "Dry-run output contains expected sections"

    output=$($BINARY install --dry-run 2>&1) || true

    if echo "$output" | grep -q "dry-run"; then
        log_pass "Output contains 'dry-run' marker"
    else
        log_fail "Output missing 'dry-run' marker"
    fi
}

test_dry_run_with_agent_flag() {
    log_test "Dry-run with --agent flag"

    output=$($BINARY install --agent claude-code --dry-run 2>&1) || true

    if echo "$output" | grep -qi "claude-code"; then
        log_pass "Dry-run output shows claude-code agent"
    else
        log_fail "Dry-run output missing claude-code agent"
    fi
}

test_dry_run_with_component_flag() {
    log_test "Dry-run with --component flag"

    output=$($BINARY install --agent opencode --component permissions --dry-run 2>&1) || true

    if echo "$output" | grep -qi "permissions"; then
        log_pass "Dry-run output shows permissions component"
    else
        log_fail "Dry-run output missing permissions component"
    fi
}

test_dry_run_platform_detection() {
    log_test "Dry-run shows platform decision"

    output=$($BINARY install --dry-run 2>&1) || true

    if echo "$output" | grep -qi "Platform decision"; then
        log_pass "Platform decision present in dry-run"
    else
        log_fail "Platform decision missing from dry-run"
    fi
}

test_dry_run_detects_linux() {
    log_test "Dry-run detects Linux OS"

    output=$($BINARY install --dry-run 2>&1) || true

    if echo "$output" | grep -q "os=linux"; then
        log_pass "Platform detected as Linux"
    else
        log_fail "Platform not detected as Linux (output: $output)"
    fi
}

test_invalid_persona_rejected() {
    log_test "Invalid persona is rejected"

    if $BINARY install --persona nonexistent --dry-run 2>&1; then
        log_fail "Invalid persona should have been rejected"
    else
        log_pass "Invalid persona correctly rejected"
    fi
}

test_invalid_component_rejected() {
    log_test "Invalid component is rejected"

    if $BINARY install --component fakecomp --dry-run 2>&1; then
        log_fail "Invalid component should have been rejected"
    else
        log_pass "Invalid component correctly rejected"
    fi
}

test_version_command() {
    log_test "Version command works"

    output=$($BINARY version 2>&1) || true

    if echo "$output" | grep -q "gentleman-ai"; then
        log_pass "Version command returns binary name"
    else
        log_fail "Version command failed: $output"
    fi
}

# ===========================================================================
# TIER 2 — Full install tests (require RUN_FULL_E2E=1)
# ===========================================================================

# --- OpenCode tests ---

test_install_opencode_permissions() {
    log_test "Install: opencode + permissions (config injection)"
    cleanup_test_env

    if $BINARY install --agent opencode --component permissions 2>&1; then
        local settings="$HOME/.config/opencode/settings.json"
        if [ -f "$settings" ]; then
            # Validate content: must have permissions.deny
            if grep -q '"permissions"' "$settings" && grep -q '"deny"' "$settings"; then
                log_pass "OpenCode settings.json has permissions config"
            else
                log_fail "OpenCode settings.json missing permissions content: $(cat "$settings")"
            fi
        else
            log_fail "OpenCode settings.json not found after install"
        fi
    else
        log_fail "Install exited with error"
    fi
}

test_install_opencode_context7() {
    log_test "Install: opencode + context7 (MCP injection)"
    cleanup_test_env

    if $BINARY install --agent opencode --component context7 2>&1; then
        local settings="$HOME/.config/opencode/settings.json"
        if [ -f "$settings" ]; then
            # Validate: must have mcpServers.context7
            if grep -q '"mcpServers"' "$settings" && grep -q '"context7"' "$settings" && grep -q 'context7-mcp' "$settings"; then
                log_pass "OpenCode settings.json has context7 MCP config"
            else
                log_fail "OpenCode settings.json missing context7 MCP: $(cat "$settings")"
            fi
        else
            log_fail "OpenCode settings.json not found after context7 install"
        fi
    else
        log_fail "Install exited with error"
    fi
}

test_install_opencode_sdd() {
    log_test "Install: opencode + sdd (command injection, no engram binary)"
    cleanup_test_env

    # SDD depends on engram in the graph. Use --component sdd,engram explicitly.
    # Engram binary install might fail (go install takes time), but SDD config
    # injection should still write the command files.
    # For this test, we only install sdd component directly with a minimal set.
    if $BINARY install --agent opencode --component sdd --component engram 2>&1; then
        local commands_dir="$HOME/.config/opencode/commands"
        if [ -d "$commands_dir" ]; then
            # Check SDD command files exist
            local cmd_count
            cmd_count=$(find "$commands_dir" -name "*.md" 2>/dev/null | wc -l)
            if [ "$cmd_count" -ge 5 ]; then
                # Validate content of one command file
                if grep -q "sdd" "$commands_dir/sdd-init.md" 2>/dev/null; then
                    log_pass "OpenCode SDD commands created ($cmd_count files)"
                else
                    log_fail "SDD command file content invalid"
                fi
            else
                log_fail "Expected >=5 SDD command files, got $cmd_count"
            fi
        else
            log_fail "OpenCode commands directory not found"
        fi
    else
        log_fail "Install exited with error"
    fi
}

test_install_opencode_persona() {
    log_test "Install: opencode + persona"
    cleanup_test_env

    if $BINARY install --agent opencode --component persona 2>&1; then
        local settings="$HOME/.config/opencode/settings.json"
        if [ -f "$settings" ]; then
            if grep -q '"persona"' "$settings"; then
                log_pass "OpenCode settings.json has persona config"
            else
                log_fail "OpenCode settings.json missing persona: $(cat "$settings")"
            fi
        else
            log_fail "OpenCode settings.json not found"
        fi
    else
        log_fail "Install exited with error"
    fi
}

# --- Claude Code tests ---

test_install_claude_code_permissions() {
    log_test "Install: claude-code + permissions"
    cleanup_test_env

    if $BINARY install --agent claude-code --component permissions 2>&1; then
        local settings="$HOME/.claude/settings.json"
        if [ -f "$settings" ]; then
            if grep -q '"permissions"' "$settings" && grep -q '"deny"' "$settings"; then
                log_pass "Claude settings.json has permissions config"
            else
                log_fail "Claude settings.json missing permissions: $(cat "$settings")"
            fi
        else
            log_fail "Claude settings.json not found"
        fi
    else
        log_fail "Install exited with error"
    fi
}

test_install_claude_code_sdd() {
    log_test "Install: claude-code + sdd (CLAUDE.md injection)"
    cleanup_test_env

    if $BINARY install --agent claude-code --component sdd --component engram 2>&1; then
        local claude_md="$HOME/.claude/CLAUDE.md"
        if [ -f "$claude_md" ]; then
            if grep -q "SDD" "$claude_md" && grep -q "sdd-init" "$claude_md"; then
                log_pass "CLAUDE.md has SDD orchestrator config"
            else
                log_fail "CLAUDE.md missing SDD content: $(cat "$claude_md")"
            fi
        else
            log_fail "CLAUDE.md not found"
        fi
    else
        log_fail "Install exited with error"
    fi
}

test_install_claude_code_context7() {
    log_test "Install: claude-code + context7 (MCP JSON injection)"
    cleanup_test_env

    if $BINARY install --agent claude-code --component context7 2>&1; then
        local mcp_file="$HOME/.claude/mcp/context7.json"
        if [ -f "$mcp_file" ]; then
            if grep -q '"command"' "$mcp_file" && grep -q 'context7-mcp' "$mcp_file"; then
                log_pass "Claude MCP context7.json created with correct content"
            else
                log_fail "Claude MCP context7.json has wrong content: $(cat "$mcp_file")"
            fi
        else
            log_fail "Claude MCP context7.json not found"
        fi
    else
        log_fail "Install exited with error"
    fi
}

# --- Both agents test ---

test_install_both_agents_permissions() {
    log_test "Install: both agents + permissions"
    cleanup_test_env

    if $BINARY install --agent opencode --agent claude-code --component permissions 2>&1; then
        local oc_settings="$HOME/.config/opencode/settings.json"
        local cc_settings="$HOME/.claude/settings.json"
        local ok=true

        if [ ! -f "$oc_settings" ]; then
            log_fail "OpenCode settings.json not found"
            ok=false
        fi
        if [ ! -f "$cc_settings" ]; then
            log_fail "Claude settings.json not found"
            ok=false
        fi

        if $ok; then
            if grep -q '"permissions"' "$oc_settings" && grep -q '"permissions"' "$cc_settings"; then
                log_pass "Both agents have permissions config"
            else
                log_fail "One or both agents missing permissions config"
            fi
        fi
    else
        log_fail "Install exited with error"
    fi
}

# ===========================================================================
# TIER 3 — Backup / restore tests (require RUN_BACKUP_TESTS=1)
# ===========================================================================

test_backup_created_on_install() {
    log_test "Backup snapshot created during install"
    cleanup_test_env
    setup_fake_configs

    if $BINARY install --agent opencode --component permissions 2>&1; then
        backup_count=$(find "$HOME/.gentleman-ai-installer/backups" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
        if [ "$backup_count" -gt 0 ]; then
            log_pass "Backup directory created ($backup_count snapshots)"
        else
            log_fail "No backup directory found"
        fi
    else
        log_fail "Install with backup failed"
    fi
}

test_backup_contains_original_files() {
    log_test "Backup snapshot contains original config files"
    cleanup_test_env
    setup_fake_configs

    if $BINARY install --agent opencode --component permissions 2>&1; then
        # Find the most recent backup snapshot
        latest_backup=$(find "$HOME/.gentleman-ai-installer/backups" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort | tail -1)
        if [ -n "$latest_backup" ]; then
            # Check that at least one file was backed up
            file_count=$(find "$latest_backup" -type f 2>/dev/null | wc -l)
            if [ "$file_count" -gt 0 ]; then
                log_pass "Backup contains $file_count file(s)"
            else
                log_fail "Backup directory is empty"
            fi
        else
            log_fail "No backup snapshot directory found"
        fi
    else
        log_fail "Install for backup test failed"
    fi
}

test_idempotent_install() {
    log_test "Idempotent: running install twice produces same result"
    cleanup_test_env

    $BINARY install --agent opencode --component permissions 2>&1 || true
    local first_content
    first_content=$(cat "$HOME/.config/opencode/settings.json" 2>/dev/null)

    $BINARY install --agent opencode --component permissions 2>&1 || true
    local second_content
    second_content=$(cat "$HOME/.config/opencode/settings.json" 2>/dev/null)

    if [ "$first_content" = "$second_content" ] && [ -n "$first_content" ]; then
        log_pass "Idempotent: same config after two runs"
    else
        log_fail "Config changed between runs"
    fi
}

# ===========================================================================
# Test execution
# ===========================================================================

log_info "=== Tier 1: Basic binary & dry-run tests ==="
test_binary_exists
test_binary_runs
test_dry_run_output_format
test_dry_run_with_agent_flag
test_dry_run_with_component_flag
test_dry_run_platform_detection
test_dry_run_detects_linux
test_invalid_persona_rejected
test_invalid_component_rejected
test_version_command

if [ "${RUN_FULL_E2E:-0}" = "1" ]; then
    log_info ""
    log_info "=== Tier 2: Full install tests ==="
    test_install_opencode_permissions
    test_install_opencode_context7
    test_install_opencode_sdd
    test_install_opencode_persona
    test_install_claude_code_permissions
    test_install_claude_code_sdd
    test_install_claude_code_context7
    test_install_both_agents_permissions
else
    log_skip "Tier 2 tests (set RUN_FULL_E2E=1 to enable)"
fi

if [ "${RUN_BACKUP_TESTS:-0}" = "1" ]; then
    log_info ""
    log_info "=== Tier 3: Backup/restore tests ==="
    test_backup_created_on_install
    test_backup_contains_original_files
    test_idempotent_install
else
    log_skip "Tier 3 tests (set RUN_BACKUP_TESTS=1 to enable)"
fi

# ---------------------------------------------------------------------------
# Summary & exit
# ---------------------------------------------------------------------------
print_summary

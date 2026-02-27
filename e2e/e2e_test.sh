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

# ===========================================================================
# TIER 2 — Full install tests (require RUN_FULL_E2E=1)
# ===========================================================================

test_install_opencode_permissions() {
    log_test "Install: opencode + permissions"
    cleanup_test_env

    if $BINARY install --agent opencode --component permissions 2>&1; then
        if [ -f "$HOME/.config/opencode/settings.json" ]; then
            log_pass "OpenCode settings.json created"
        else
            log_fail "OpenCode settings.json not found after install"
        fi
    else
        log_fail "Install exited with error"
    fi
}

test_install_claude_code_persona() {
    log_test "Install: claude-code + persona"
    cleanup_test_env

    if $BINARY install --agent claude-code --component persona 2>&1; then
        if [ -f "$HOME/.claude/settings.json" ]; then
            log_pass "Claude Code settings.json created"
        else
            log_fail "Claude Code settings.json not found after install"
        fi
    else
        log_fail "Install exited with error"
    fi
}

test_install_opencode_context7() {
    log_test "Install: opencode + context7"
    cleanup_test_env

    if $BINARY install --agent opencode --component context7 2>&1; then
        if [ -f "$HOME/.config/opencode/settings.json" ]; then
            log_pass "OpenCode settings.json created with context7"
        else
            log_fail "OpenCode settings.json not found after context7 install"
        fi
    else
        log_fail "Install exited with error"
    fi
}

test_install_opencode_sdd() {
    log_test "Install: opencode + sdd"
    cleanup_test_env

    if $BINARY install --agent opencode --component sdd 2>&1; then
        if [ -d "$HOME/.config/opencode/commands" ]; then
            log_pass "OpenCode SDD commands directory created"
        else
            log_fail "OpenCode SDD commands directory not found"
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

if [ "${RUN_FULL_E2E:-0}" = "1" ]; then
    log_info ""
    log_info "=== Tier 2: Full install tests ==="
    test_install_opencode_permissions
    test_install_claude_code_persona
    test_install_opencode_context7
    test_install_opencode_sdd
else
    log_skip "Tier 2 tests (set RUN_FULL_E2E=1 to enable)"
fi

if [ "${RUN_BACKUP_TESTS:-0}" = "1" ]; then
    log_info ""
    log_info "=== Tier 3: Backup/restore tests ==="
    test_backup_created_on_install
    test_backup_contains_original_files
else
    log_skip "Tier 3 tests (set RUN_BACKUP_TESTS=1 to enable)"
fi

# ---------------------------------------------------------------------------
# Summary & exit
# ---------------------------------------------------------------------------
print_summary

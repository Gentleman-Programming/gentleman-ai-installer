#!/usr/bin/env bash
# lib.sh — shared test helpers for gentleman-ai E2E tests
# Sourced by e2e_test.sh; never executed directly.
set -euo pipefail

# ---------------------------------------------------------------------------
# Colors
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Counters
# ---------------------------------------------------------------------------
PASSED=0
FAILED=0
SKIPPED=0

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------
log_test()  { printf "${YELLOW}[TEST]${NC}  %s\n" "$1"; }
log_pass()  { printf "${GREEN}[PASS]${NC}  %s\n" "$1"; PASSED=$((PASSED + 1)); }
log_fail()  { printf "${RED}[FAIL]${NC}  %s\n" "$1"; FAILED=$((FAILED + 1)); }
log_skip()  { printf "${BLUE}[SKIP]${NC}  %s\n" "$1"; SKIPPED=$((SKIPPED + 1)); }
log_info()  { printf "${BLUE}[INFO]${NC}  %s\n" "$1"; }

# ---------------------------------------------------------------------------
# Binary resolution
# ---------------------------------------------------------------------------
# The binary should be built and placed at /usr/local/bin/gentleman-ai inside
# the Docker container. If not found, fall back to $HOME/gentleman-ai or the
# current directory.
resolve_binary() {
    if command -v gentleman-ai >/dev/null 2>&1; then
        echo "gentleman-ai"
    elif [ -x "$HOME/gentleman-ai" ]; then
        echo "$HOME/gentleman-ai"
    elif [ -x "./gentleman-ai" ]; then
        echo "./gentleman-ai"
    else
        echo ""
    fi
}

# ---------------------------------------------------------------------------
# Cleanup helpers
# ---------------------------------------------------------------------------

# cleanup_test_env — reset filesystem state between tests.
# Removes config dirs and files that the installer writes.
cleanup_test_env() {
    rm -rf "$HOME/.config/opencode" 2>/dev/null || true
    rm -rf "$HOME/.claude" 2>/dev/null || true
    rm -rf "$HOME/.gentleman-ai-installer" 2>/dev/null || true
    mkdir -p "$HOME/.config"
}

# setup_fake_configs — seed fake config files so backup tests have something
# to snapshot and restore.
setup_fake_configs() {
    mkdir -p "$HOME/.config/opencode"
    echo '{"fake-settings": true}' > "$HOME/.config/opencode/settings.json"

    mkdir -p "$HOME/.claude"
    echo '# Fake CLAUDE.md' > "$HOME/.claude/CLAUDE.md"
}

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
print_summary() {
    echo ""
    echo "========================================"
    printf "  ${GREEN}PASSED${NC}: %d\n" "$PASSED"
    printf "  ${RED}FAILED${NC}: %d\n" "$FAILED"
    printf "  ${BLUE}SKIPPED${NC}: %d\n" "$SKIPPED"
    echo "  TOTAL : $((PASSED + FAILED + SKIPPED))"
    echo "========================================"

    if [ "$FAILED" -gt 0 ]; then
        printf "\n%bSome tests failed.%b\n" "$RED" "$NC"
        return 1
    fi

    printf "\n%bAll tests passed.%b\n" "$GREEN" "$NC"
    return 0
}

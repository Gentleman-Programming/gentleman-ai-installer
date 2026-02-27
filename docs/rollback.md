# Rollback Guide

The installer snapshot system stores a manifest and copied files before configuration writes. Rollback behavior is platform-agnostic — it works the same on macOS and Linux.

## Snapshot contents

- `manifest.json` with all tracked entries.
- `files/...` tree with per-file copies.
- For paths that did not exist before install, manifest tracks `existed=false`.

## Restore behavior

- If `existed=true`, restore copies back to original paths.
- If `existed=false`, remove newly created files.
- Restore is atomic per file write.

This applies regardless of the detected platform. The snapshot captures file state, not platform-specific package state — package installs (brew/apt/pacman) are not rolled back by the snapshot system.

## If verification fails

1. Review failed checks in verification report.
2. Restore from latest snapshot manifest.
3. Re-run install with `--dry-run` to validate plan.
4. Re-run install after fixing external dependencies (for example: missing binary or unavailable service).

## What rollback does NOT cover

- Packages installed via `brew install`, `apt-get install`, or `pacman -S` are not uninstalled during rollback. The snapshot system handles configuration files only.
- If you need to undo a package install, use your platform's package manager directly (e.g., `brew uninstall`, `sudo apt-get remove`, `sudo pacman -R`).

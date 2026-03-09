#!/usr/bin/env bash

# ============================================================================
# Gentleman Guardian Angel - PR Mode Functions
# ============================================================================
# Handles PR-scoped code review:
# - detect_base_branch(): Auto-detect main/master/develop
# - get_pr_range(): Build the git range for PR comparison
# - get_pr_files(): Get files changed in the PR range
# - get_pr_diff(): Get the diff for the PR range
# - validate_pr_mode_flags(): Validate --pr-mode and --diff-only flags
# - build_pr_prompt(): Build review prompt for PR mode
# ============================================================================

# ============================================================================
# Base Branch Detection
# ============================================================================

# Detect the base branch (main, master, or develop)
# Preference order: main > master > develop
# Returns: the branch name on stdout
detect_base_branch() {
  local branches
  branches=$(git branch 2>/dev/null)

  if [[ -z "$branches" ]]; then
    echo "Error: Could not detect base branch (not a git repo?)" >&2
    return 1
  fi

  # Check in priority order: main > master > develop
  for candidate in main master develop; do
    if echo "$branches" | grep -qw "$candidate"; then
      echo "$candidate"
      return 0
    fi
  done

  echo "Error: Could not detect base branch. No main, master, or develop branch found." >&2
  echo "Set PR_BASE_BRANCH in your .gga config to specify the base branch." >&2
  return 1
}

# ============================================================================
# PR Range
# ============================================================================

# Get the git range for PR comparison
# Usage: get_pr_range <pr_base_branch_override>
# If override is empty, auto-detect base branch
get_pr_range() {
  local pr_base_branch="$1"

  if [[ -n "$pr_base_branch" ]]; then
    echo "${pr_base_branch}...HEAD"
    return 0
  fi

  local base
  if ! base=$(detect_base_branch); then
    return 1
  fi

  echo "${base}...HEAD"
  return 0
}

# ============================================================================
# PR Files
# ============================================================================

# Get files changed in the PR range, filtered by patterns
# Usage: get_pr_files <range> <file_patterns> <exclude_patterns>
get_pr_files() {
  local range="$1"
  local patterns="$2"
  local excludes="$3"

  # Get all changed files in the range
  local changed
  changed=$(git diff --name-only --diff-filter=ACM "$range" 2>/dev/null)

  if [[ -z "$changed" ]]; then
    return
  fi

  # Convert comma-separated patterns to array
  IFS=',' read -ra PATTERN_ARRAY <<< "$patterns"
  IFS=',' read -ra EXCLUDE_ARRAY <<< "$excludes"

  # Filter files
  echo "$changed" | while IFS= read -r file; do
    local match=false
    local excluded=false

    # Check if file matches any include pattern
    for pattern in "${PATTERN_ARRAY[@]}"; do
      pattern=$(echo "$pattern" | xargs) # trim whitespace
      if [[ "$pattern" == "*" ]]; then
        match=true
        break
      elif [[ "$pattern" == \** ]]; then
        local suffix="${pattern#\*}"
        if [[ "$file" == *"$suffix" ]]; then
          match=true
          break
        fi
      else
        # shellcheck disable=SC2053
        if [[ "$file" == $pattern ]] || [[ "$(basename "$file")" == $pattern ]]; then
          match=true
          break
        fi
      fi
    done

    # Check if file matches any exclude pattern
    if [[ "$match" == true && -n "$excludes" ]]; then
      for pattern in "${EXCLUDE_ARRAY[@]}"; do
        pattern=$(echo "$pattern" | xargs) # trim whitespace
        if [[ "$pattern" == \** ]]; then
          local suffix="${pattern#\*}"
          if [[ "$file" == *"$suffix" ]]; then
            excluded=true
            break
          fi
        else
          # shellcheck disable=SC2053
          if [[ "$file" == $pattern ]] || [[ "$(basename "$file")" == $pattern ]]; then
            excluded=true
            break
          fi
        fi
      done
    fi

    if [[ "$match" == true && "$excluded" == false ]]; then
      # Only include if file still exists (wasn't deleted)
      # GGA_SKIP_FILE_CHECK allows unit tests to skip this check
      if [[ -n "${GGA_SKIP_FILE_CHECK:-}" ]] || [[ -f "$file" ]]; then
        echo "$file"
      fi
    fi
  done
}

# ============================================================================
# PR Diff
# ============================================================================

# Get the unified diff for the PR range
# Usage: get_pr_diff <range>
get_pr_diff() {
  local range="$1"
  git diff "$range" 2>/dev/null
}

# ============================================================================
# Flag Validation
# ============================================================================

# Validate PR mode flags
# Usage: validate_pr_mode_flags <pr_mode> <diff_only>
validate_pr_mode_flags() {
  local pr_mode="$1"
  local diff_only="$2"

  if [[ "$diff_only" == "true" && "$pr_mode" != "true" ]]; then
    echo "Error: --diff-only can only be used with --pr-mode" >&2
    return 1
  fi

  return 0
}

# ============================================================================
# PR Prompt Building
# ============================================================================

# Build the review prompt for PR mode
# Usage: build_pr_prompt <rules> <files> <diff_only> <diff_content> <base_branch>
build_pr_prompt() {
  local rules="$1"
  local files="$2"
  local diff_only="$3"
  local diff_content="$4"
  local base_branch="$5"

  cat << EOF
You are a code reviewer analyzing a pull request against the ${base_branch} branch.

=== CODING STANDARDS ===
${rules}
=== END CODING STANDARDS ===

=== PR CONTEXT ===
This is a pull request review. The following files were changed in this PR (compared to ${base_branch}).
=== END PR CONTEXT ===
EOF

  if [[ "$diff_only" == "true" && -n "$diff_content" ]]; then
    cat << EOF

=== PR DIFF ===
${diff_content}
=== END PR DIFF ===

=== FILES (complete content for context) ===
EOF
  else
    cat << 'EOF'

=== FILES TO REVIEW ===
EOF
  fi

  # Add file contents
  while IFS= read -r file; do
    if [[ -n "$file" ]]; then
      echo ""
      echo "--- FILE: $file ---"
      if [[ -f "$file" ]]; then
        cat "$file"
      fi
    fi
  done <<< "$files"

  cat << 'EOF'

=== END FILES ===

**IMPORTANT: Your response MUST include one of these lines near the beginning:**
STATUS: PASSED
STATUS: FAILED

**If FAILED:** List each violation with:
- File name
- Line number (if applicable)
- Rule violated
- Description of the issue

**If PASSED:** Confirm all files comply with the coding standards.

**Begin with STATUS:**
EOF
}

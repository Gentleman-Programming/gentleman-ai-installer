# Code Review Rules

## General
REJECT if:
- Hardcoded secrets or credentials
- `any` type (TypeScript) or missing type hints (Python)
- Empty catch blocks (silent error handling)
- Code duplication (violates DRY)
- `console.log` / `print()` in production code

## TypeScript/React
REJECT if:
- `import * as React` → use `import { useState }` (named imports)
- `var()` or hex colors in className → use Tailwind utilities
- `useMemo`/`useCallback` without justification (React 19 Compiler handles this)
- Missing `"use client"` in client components

PREFER:
- `cn()` for conditional class merging
- Semantic HTML over divs
- Named exports over default exports

## Python
REJECT if:
- Missing type hints on public functions
- Bare `except:` without specific exception
- `print()` instead of `logger`

## Go
REJECT if:
- Exported functions without doc comments
- Ignored errors (no `_ = err`)
- Naked returns in long functions

## Response Format
FIRST LINE must be exactly:
STATUS: PASSED
or
STATUS: FAILED

If FAILED, list: `file:line - rule violated - issue`

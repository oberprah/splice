# Go Version Research

Researched: 2025-12-26

## Current Go Versions

- **Go 1.25.5** (latest) - released December 2, 2025
- **Go 1.25.2** - released October 7, 2025 (security patch)
- **Go 1.24.11** - latest in 1.24 series

Go supports the two most recent major versions (currently 1.25 and 1.24).

## GitHub Actions setup-go Support

The `actions/setup-go@v6` action:

- Supports all current Go versions including 1.25.x
- Can auto-detect version from `go.mod`, `go.work`, `.go-version`, or `.tool-versions`
- Supports specific versions (e.g., `1.25.2`) or SemVer ranges (e.g., `^1.25.0`)
- Version numbers should be quoted in YAML (e.g., `go-version: '1.25'`)

## Recommendation

Use `go-version-file: 'go.mod'` to automatically use whatever version is specified in the project's go.mod file. This keeps the CI in sync with the project without manual updates.

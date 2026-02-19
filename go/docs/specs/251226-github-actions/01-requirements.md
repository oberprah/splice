# Requirements: GitHub Actions Pipeline

## Problem Statement

The Splice repository lacks automated CI/CD. Without it, broken builds, test failures, and code quality issues can be merged undetected. Dependencies also become stale over time without proactive updates.

## Goals

1. **Catch regressions early** - Validate builds, tests, and code quality on every PR and push to main
2. **Keep dependencies fresh** - Automatically create PRs when dependencies have updates

## Non-Goals

- Automated releases/deployments
- Test coverage reporting or badges
- Auto-merging dependency updates

## Requirements

### CI Pipeline

| Requirement | Details |
|-------------|---------|
| **Trigger** | On pull requests and pushes to `main` branch |
| **Build** | Run `go build` to verify compilation |
| **Test** | Run `go test ./...` to verify all tests pass |
| **Lint** | Run `golangci-lint` for static analysis |
| **Go version** | Use version from `go.mod` (currently 1.25.2) |

### Dependency Updates

| Requirement | Details |
|-------------|---------|
| **Tool** | GitHub Dependabot |
| **Behavior** | Create PRs when Go module dependencies have updates |
| **Review** | Manual review and merge by maintainer |

## User Impact

- Contributors get immediate feedback on PRs before review
- Maintainers can trust that merged code builds and passes tests
- Dependencies stay current through regular update PRs

## Research

- [go-versions.md](research/go-versions.md) - Go version compatibility with GitHub Actions

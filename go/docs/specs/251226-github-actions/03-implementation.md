# Implementation Plan: GitHub Actions Pipeline

## Steps

- [x] Step 1: Create CI workflow, Dependabot config, and golangci-lint config files
- [x] Step 2: Use Go tool directive for golangci-lint and fix lint issues
- [x] Validation: Verify all tests pass, build succeeds, and lint runs clean

## Progress

### Step 1: Create configuration files
Status: ✅ Complete
Commits: 8003ecf
Notes: Created all three files according to design specifications:
- `.github/workflows/ci.yml` - CI pipeline with parallel lint, test, build jobs
- `.github/dependabot.yml` - Weekly Go module updates, grouped into single PR
- `.golangci.yml` - Minimal linter config with gofmt and goimports enabled

Key decisions: Used latest stable action versions (checkout@v5, setup-go@v6, golangci-lint-action@v9), set minimal read-only permissions, followed Unix newline conventions

### Step 2: Use Go tool directive and fix lint issues
Status: ✅ Complete
Commits: 9352b4f
Notes:
- Added golangci-lint v1.64.8 as tool dependency via `go get -tool`
- Updated CI workflow to use `go tool golangci-lint run` instead of golangci-lint-action
- Fixed lint issues:
  - Formatting issues across codebase (gofmt)
  - `gosimple S1029` in diff_view.go: Changed to direct string ranging
  - `staticcheck SA1019` in diff_view.go: Removed deprecated `syntaxStyle.Copy()` call
  - `staticcheck SA4017` in git.go: Removed dead code with unused return value

**Design deviation**: Used Go's native tool directive instead of golangci-lint-action for better version pinning, reproducibility, and alignment between local dev and CI environments

### Validation
Status: ✅ Complete
Notes: All checks pass locally

## Discoveries

- **Go 1.24+ tool directive**: Discovered that Go 1.24+ has native support for tracking tool dependencies in go.mod via the `tool` directive. This is preferred over manual installation or the old `tools.go` pattern.
- **golangci-lint installation**: The golangci-lint docs recommend against `go install` due to untested binaries, but the `go tool` approach downloads pre-built packages from the module cache, which is acceptable.
- **Design improvement**: Using `go tool golangci-lint run` in CI aligns local dev and CI environments, ensures version consistency, and eliminates need for third-party GitHub Actions.

## Verification

- [x] All tests pass (`go test ./...`)
- [x] Build succeeds (`go build ./...`)
- [x] Lint runs clean (`go tool golangci-lint run`)
- [x] Requirements verified against 01-requirements.md
- [x] All changes committed (implementation code)

### Requirements Verification

| Requirement | Status | Implementation |
|-------------|--------|----------------|
| Trigger on PRs and push to main | ✅ | `on: push: branches: [main]` + `pull_request` |
| Run `go build` | ✅ | Build job with `go build ./...` |
| Run `go test ./...` | ✅ | Test job with `go test ./...` |
| Run golangci-lint | ✅ | Lint job with `go tool golangci-lint run` |
| Use Go version from go.mod | ✅ | `go-version-file: 'go.mod'` |
| Dependabot for Go modules | ✅ | `.github/dependabot.yml` configured |
| Weekly dependency updates | ✅ | `schedule: interval: weekly, day: monday` |
| Manual review of updates | ✅ | No auto-merge configured |

## Summary

Implementation complete. The GitHub Actions CI pipeline is ready for testing.

**Files created:**
- `.github/workflows/ci.yml` - CI pipeline with parallel lint, test, build jobs
- `.github/dependabot.yml` - Weekly grouped dependency update PRs
- `.golangci.yml` - Minimal linter configuration

**Key improvement over design:**
Used Go's native tool directive (`go get -tool`) for golangci-lint instead of the golangci-lint-action. This ensures version consistency between local development and CI, with the linter version pinned in `go.mod`.

**Testing instructions:**
1. Push this branch to GitHub
2. Open a PR to `main`
3. Verify all three CI jobs (lint, test, build) run and pass
4. Merge to main and verify CI runs on the push event

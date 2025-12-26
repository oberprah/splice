# Design: GitHub Actions Pipeline

## Executive Summary

This design implements a CI pipeline using GitHub Actions with three parallel jobs: lint, test, and build. Each job runs independently for faster feedback. The pipeline triggers on PRs and pushes to main, using the Go version from `go.mod` to stay in sync with the project.

For dependency updates, Dependabot will check weekly for Go module updates and create grouped PRs for manual review.

The implementation requires three files: a CI workflow, a Dependabot configuration, and a golangci-lint configuration. Any existing lint issues will be fixed as part of this work (or split into a separate PR if extensive).

## Context & Problem Statement

The Splice repository lacks automated CI/CD. Without it, broken builds, test failures, and code quality issues can be merged undetected. Dependencies also become stale without proactive updates.

**Scope:** This design covers CI pipeline and dependency updates. It does not cover automated releases, coverage reporting, or auto-merging.

## Current State

- No `.github/` directory exists
- No CI/CD configuration
- Tests exist and pass locally (~4 seconds total runtime)
- No linting configuration
- Go version 1.25.2 specified in `go.mod`

## Solution

### Pipeline Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    GitHub Event                         │
│         (PR opened/updated or push to main)             │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
   ┌─────────┐        ┌─────────┐        ┌─────────┐
   │  Lint   │        │  Test   │        │  Build  │
   │         │        │         │        │         │
   │ golangci│        │ go test │        │go build │
   │  -lint  │        │  ./...  │        │  ./...  │
   └─────────┘        └─────────┘        └─────────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           ▼
                    ┌─────────────┐
                    │   Status    │
                    │  (pass/fail)│
                    └─────────────┘
```

> **Decision:** Run lint, test, and build as separate parallel jobs rather than sequential steps in one job. This provides faster feedback — a lint failure is visible immediately without waiting for tests, and all three run simultaneously.

### Trigger Configuration

```yaml
on:
  push:
    branches: [main]
  pull_request:
```

> **Decision:** Trigger on pushes to `main` AND pull requests (not `on: [push, pull_request]` which causes double runs). This runs once on PRs and once when merging to main.

### Go Version Strategy

> **Decision:** Use `go-version-file: 'go.mod'` instead of hardcoding a version. This keeps CI automatically in sync with the project without manual updates when Go version changes.

### Linting Configuration

> **Decision:** Use golangci-lint with a minimal configuration that enables useful linters beyond the defaults. The configuration will be a separate `.golangci.yml` file rather than inline in the workflow for maintainability.

Selected linters:
- **errcheck** — unchecked errors (default)
- **govet** — suspicious constructs (default)
- **staticcheck** — static analysis (default)
- **unused** — unused code (default)
- **gofmt** — formatting
- **goimports** — import ordering

> **Decision:** Run lint on all code (no `--only-new-issues` flag). Any existing lint issues will be fixed as part of this implementation. If the cleanup proves extensive, it will be split into a separate PR.

### Dependabot Configuration

```
┌─────────────────────────────────────────────┐
│              Dependabot                      │
│  Weekly check (Mondays) for Go modules      │
└─────────────────────────────────────────────┘
                    │
                    ▼
            ┌───────────────┐
            │  Grouped PR   │
            │  (all deps)   │
            └───────────────┘
                    │
                    ▼
            ┌───────────────┐
            │ Manual Review │
            │   & Merge     │
            └───────────────┘
```

> **Decision:** Group all Go dependency updates into a single PR rather than one PR per dependency. This reduces PR noise significantly for a project with many indirect dependencies.

> **Decision:** Schedule updates weekly on Mondays. Daily is too frequent for this project's needs; weekly provides regular freshness without overwhelming with PRs.

### Files to Create

| File | Purpose |
|------|---------|
| `.github/workflows/ci.yml` | CI pipeline definition |
| `.github/dependabot.yml` | Dependabot configuration |
| `.golangci.yml` | Linter configuration |

### CI Workflow Structure

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
        with:
          go-version-file: 'go.mod'
      - uses: golangci/golangci-lint-action@v9
        with:
          version: v2

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
        with:
          go-version-file: 'go.mod'
      - run: go test ./...

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v6
        with:
          go-version-file: 'go.mod'
      - run: go build ./...
```

### Dependabot Configuration Structure

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
    groups:
      go-dependencies:
        patterns:
          - "*"
```

### golangci-lint Configuration Structure

```yaml
linters:
  enable:
    - gofmt
    - goimports

run:
  timeout: 5m
```

> **Decision:** Keep configuration minimal — only enable linters beyond defaults that provide clear value. The defaults (errcheck, govet, staticcheck, unused) are already valuable. Adding gofmt and goimports catches formatting issues.

## Open Questions

None. The design is straightforward and follows established patterns. All decisions have been made and documented above.

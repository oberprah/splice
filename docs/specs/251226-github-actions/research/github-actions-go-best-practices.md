# GitHub Actions Best Practices for Go Projects

Researched: 2025-12-26

## Key Findings

### 1. Workflow Structure

**Separate Jobs for Different Concerns:**
Run linting, testing, and building in separate jobs for parallel execution:

```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
  test:
    runs-on: ubuntu-latest
  build:
    runs-on: ubuntu-latest
```

### 2. Caching

**Built-in Caching with setup-go:**
Since v4, `actions/setup-go` includes automatic caching of `GOCACHE` and `GOMODCACHE` using `go.sum` as the cache key. Cache is enabled by default.

Add `go mod download` early to ensure modules are cached:

```yaml
- name: Download dependencies
  run: go mod download
```

### 3. Trigger Configuration

**Avoid Double Execution:**
Don't use `on: [push, pull_request]` — it causes double runs. Instead:

```yaml
on:
  push:
    branches: [main]
  pull_request:
```

This runs on pushes to main OR when a PR is opened/updated.

### 4. golangci-lint Integration

**Official Action (v9):**

```yaml
- uses: golangci/golangci-lint-action@v9
  with:
    version: v2
```

Key options:
- `only-new-issues: true` — focus on changes in PRs
- Built-in caching for faster runs
- Line-attached annotations in PR view

### 5. Current Action Versions (December 2025)

- `actions/checkout@v5`
- `actions/setup-go@v6`
- `golangci/golangci-lint-action@v9`
- `actions/upload-artifact@v4`

## Sources

- [Building and testing Go - GitHub Docs](https://docs.github.com/en/actions/use-cases-and-examples/building-and-testing/building-and-testing-go)
- [golangci/golangci-lint-action](https://github.com/golangci/golangci-lint-action)
- [actions/setup-go](https://github.com/actions/setup-go)

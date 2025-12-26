# Dependabot Configuration for Go Modules

Researched: 2025-12-26

## File Location

`.github/dependabot.yml` at repository root.

## Basic Structure

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
```

## Required Fields

- `version: 2` — must be version 2
- `package-ecosystem: "gomod"` — not "go" or "golang"
- `directory: "/"` — location of go.mod
- `schedule` — update frequency

## Common Options

| Option | Purpose |
|--------|---------|
| `schedule.interval` | `daily`, `weekly`, or `monthly` |
| `schedule.day` | Day for weekly (e.g., `monday`) |
| `open-pull-requests-limit` | Max concurrent PRs (default: 5) |
| `reviewers` | Auto-assign reviewers |
| `labels` | Custom labels |
| `groups` | Group related updates |

## Grouping Dependencies

Reduces PR noise by combining related updates:

```yaml
groups:
  go-deps:
    patterns:
      - "*"
```

## Gotchas

1. **Vendor Support** — Automatic if using `go mod vendor`
2. **Labels** — Always applies `dependencies` + `gomod` labels automatically
3. **Immediate Check** — Adding config triggers immediate version check

## Sources

- [Dependabot options reference - GitHub Docs](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file)
- [Configuring Dependabot version updates - GitHub Docs](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuring-dependabot-version-updates)

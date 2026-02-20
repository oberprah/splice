# CLI Parsing Research

Research findings for implementing `splice diff <spec>` subcommand parsing.

## Recommended Approach: Manual Parsing

Given Splice's philosophy of lean, focused interfaces, manual argument parsing is recommended over heavy frameworks.

### Why Manual Parsing?

**Options Evaluated:**
- **Cobra**: Heavy framework, overkill for 2 subcommands
- **Flag package**: Too verbose for subcommand pattern with arbitrary args
- **Manual parsing**: ~50 lines of code, zero new dependencies, full control

**Advantages:**
- Aligns with Splice's "lean" philosophy
- Only 2 commands needed (log, diff)
- No complex subcommand nesting
- Zero new dependencies
- Clear, testable code

## Implementation Pattern

### Argument Parsing

```go
// main.go - Parse arguments and route to appropriate initialization
func parseArgs() (cmd string, args []string) {
    if len(os.Args) < 2 {
        return "log", []string{} // Default: splice -> log view
    }

    firstArg := os.Args[1]
    if firstArg == "diff" {
        return "diff", os.Args[2:] // Everything after "diff"
    }

    return "log", []string{} // Unknown -> log view or error
}
```

### Main Entry Point

```go
func main() {
    cmd, args := parseArgs()

    var initialState core.State
    if cmd == "diff" {
        spec, err := parseDiffSpec(args)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            os.Exit(1)
        }
        initialState = createDirectDiffState(spec)
    } else {
        initialState = loading.State{}
    }

    model := app.NewModel(app.WithInitialState(initialState))
    p := tea.NewProgram(model, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Diff Spec Parsing

```go
func parseDiffSpec(args []string) (DiffSpec, error) {
    if len(args) == 0 {
        // splice diff -> unstaged changes
        return DiffSpec{Type: TypeUnstaged}, nil
    }

    firstArg := args[0]

    // Handle special flags
    switch firstArg {
    case "--staged", "--cached":
        if len(args) > 1 {
            return DiffSpec{}, fmt.Errorf("--staged does not accept additional arguments")
        }
        return DiffSpec{Type: TypeStaged}, nil
    case "HEAD":
        if len(args) > 1 {
            return DiffSpec{}, fmt.Errorf("unexpected arguments after HEAD")
        }
        return DiffSpec{Type: TypeAllUncommitted}, nil
    default:
        // Assume it's a commit spec: main..feature, HEAD~5..HEAD, etc.
        if len(args) > 1 {
            return DiffSpec{}, fmt.Errorf("unexpected arguments: %v", args[1:])
        }

        if !isValidDiffSpec(firstArg) {
            return DiffSpec{}, fmt.Errorf("invalid diff spec: %q", firstArg)
        }

        return DiffSpec{Type: TypeCommitRange, RawSpec: firstArg}, nil
    }
}

func isValidDiffSpec(spec string) bool {
    // Basic validation - disallow obvious junk
    return !strings.Contains(spec, " ") &&
           !strings.Contains(spec, ";") &&
           !strings.Contains(spec, "|")
}
```

## Command Routing

```
splice              -> LoadingState (fetch commits, show log)
splice diff         -> DirectDiffState (show unstaged changes)
splice diff --staged -> DirectDiffState (show staged changes)
splice diff HEAD    -> DirectDiffState (show all uncommitted)
splice diff main..feature -> DirectDiffState (show branch comparison)
```

## Error Handling

Validate early, fail fast in main.go before entering TUI:

```go
func validateDiffSpec(spec DiffSpec) error {
    // Try: git diff --quiet <spec>
    // Exit code 0 = no changes
    // Exit code 1 = has changes
    // Exit code 128+ = error (invalid spec)

    cmd := exec.Command("git", "diff", "--quiet", spec.GitArgs()...)
    err := cmd.Run()

    if err == nil {
        return fmt.Errorf("no changes in %q", spec)
    }

    if exitErr, ok := err.(*exec.ExitError); ok {
        code := exitErr.ExitCode()
        if code == 1 {
            return nil // Has changes, valid
        }
        return fmt.Errorf("invalid diff spec: %q", spec)
    }

    return err
}
```

## Testing Strategy

**Unit Tests:**
- `parseDiffSpec()` - valid/invalid specs
- `isValidDiffSpec()` - format validation
- Edge cases: empty args, multiple args, special characters

**Integration Tests:**
- Real git repository with branches
- Uncommitted changes scenarios
- Invalid specs (should error gracefully)

**E2E Tests:**
- Full workflow: `splice diff main..feature` -> select file -> view diff -> exit

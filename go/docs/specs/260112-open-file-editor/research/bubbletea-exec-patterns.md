# Research: Bubbletea Suspend/Resume and External Command Execution

## Key Findings

Bubbletea has excellent built-in support for suspending the TUI, running external commands, and resuming through the `tea.ExecProcess()` function.

## The ExecProcess Function

**API:**
```go
func ExecProcess(c *exec.Cmd, fn ExecCallback) Cmd
type ExecCallback func(error) Msg
```

**How it works:**
- Calls `ReleaseTerminal()` before running (restores original terminal state)
- Runs the external command with access to stdin/stdout
- Calls `RestoreTerminal()` after completion (reinitializes Bubbletea)
- Resets the renderer line tracking

## Recommended Pattern for Editor Integration

```go
// Message type for editor operation
type EditorFinishedMsg struct {
    err error
}

// Function to launch editor
func launchEditor(editorPath string, filePath string, lineNumber int) tea.Cmd {
    // Use "+lineNo | normal! zt" for vim/nvim to position line at top (not centered)
    cmd := exec.Command(editorPath, fmt.Sprintf("+%d | normal! zt", lineNumber), filePath)

    return tea.ExecProcess(cmd, func(err error) tea.Msg {
        return EditorFinishedMsg{err: err}
    })
}

// In Update method:
case "o": // Open in editor key
    return s, launchEditor(editorPath, filePath, lineNumber)

case EditorFinishedMsg:
    if msg.err != nil {
        // Show error to user
        return s, showError(msg.err)
    }
    // Resume with preserved state (diff view, scroll position remain intact)
    return s, nil
```

## Editor Command-Line Syntax for Line Positioning

| Editor | Syntax | Example |
|--------|--------|---------|
| vim | `+<line>` | `vim +42 file.go` |
| nvim | `+<line>` | `nvim +42 file.go` |
| vi | `+<line>` | `vi +42 file.go` |
| nano | `+<line>` | `nano +42 file.go` |
| emacs | `+<line>` | `emacs +42 file.go` |

All common terminal editors support the `+line` syntax.

### Vim Line Positioning Behavior

**Default behavior:** When vim/nvim opens with `+lineNo`, it centers the target line in the viewport (similar to the `zz` command).

**Problem:** If you want the line at the top of the diff view to appear at the top of the editor (not centered), you need to add positioning commands.

**Solution:** Use `"+lineNo | normal! zt"` to position the line at the **t**op of the screen:
```go
cmd := exec.Command(editor, fmt.Sprintf("+%d | normal! zt", lineNo), absolutePath)
```

**Vim positioning commands:**
- `zt` - Position current line at **t**op of screen
- `zz` - Position current line at center (default behavior)
- `zb` - Position current line at **b**ottom of screen

## Best Practices

1. **Always use `ExecProcess()` for interactive editors** - Handles terminal state automatically
2. **Create a typed message for the callback** - Enables proper message routing
3. **Handle the error case** - Always check `EditorFinishedMsg.err`
4. **Preserve application state** - ExecProcess doesn't affect UI state
5. **Validate before launching** - Check env vars and file existence

## Potential Pitfalls

| Pitfall | Solution |
|---------|----------|
| Forgetting to wrap with `ExecProcess()` | Always use `ExecProcess()` |
| Not handling callback errors | Always handle `EditorFinishedMsg.err` |
| Using relative file paths | Resolve to absolute path first |
| Not checking env vars | Check both `$EDITOR` and `$VISUAL` |

## References

- Official Bubbletea Docs: https://pkg.go.dev/github.com/charmbracelet/bubbletea#ExecProcess
- Exec Example: https://github.com/charmbracelet/bubbletea/tree/main/examples/exec

# Design Doc: Direct Diff Command

## Overview

Support launching directly into the diff view via CLI commands like `splice diff main..feature`. This enables quick access to specific diffs without navigating through the log view.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Abstraction | `DiffSource` enum | Matches Go implementation, cleanly handles both commit ranges and uncommitted changes |
| CLI parsing | Manual (no clap) | Keeps dependencies minimal, consistent with existing approach |
| Validation phase | Pre-TUI validation | Exit early with error message if spec is invalid, matches Go |
| Initial state | Direct to FilesView | No loading state needed since we validate before entering TUI |

## Command Examples

```bash
splice                           # Log view (default, existing behavior)
splice diff                      # Unstaged changes (working tree vs index)
splice diff --staged             # Staged changes (index vs HEAD)
splice diff HEAD                 # All uncommitted changes (working tree vs HEAD)
splice diff main..feature        # Commit range (two-dot)
splice diff main...feature       # Commit range (three-dot, merge base)
splice diff abc123               # Single commit
```

## Domain Objects

### DiffSource (new)

```rust
pub enum DiffSource {
    CommitRange(CommitRange),
    Uncommitted(UncommittedType),
}

pub enum UncommittedType {
    Unstaged,   // Working tree vs index (git diff)
    Staged,     // Index vs HEAD (git diff --staged)
    All,        // Working tree vs HEAD (git diff HEAD)
}
```

**Why an enum instead of extending CommitRange?**
- Uncommitted changes have no commit hashes
- Different git commands needed (`git diff` vs `git diff-tree`)
- Clean separation of concerns
- Future-proof for other diff sources (e.g., stash, worktree)

### DiffSpec (new, for CLI parsing)

```rust
pub struct DiffSpec {
    pub raw: Option<String>,           // For commit ranges like "main..feature"
    pub uncommitted_type: Option<UncommittedType>,
}
```

This is a parsed-but-not-yet-resolved representation from CLI args.

## File Changes

### src/cli/mod.rs (new)

```rust
pub enum Command {
    Log,
    Diff(DiffSpec),
}

pub fn parse_args(args: &[String]) -> Command;

mod diff_spec;
mod validation;
```

### src/cli/diff_spec.rs (new)

```rust
pub fn parse_diff_args(args: &[String]) -> Result<DiffSpec, String>;

fn is_valid_diff_spec(spec: &str) -> bool;
```

**Parsing rules:**
- No args → `UncommittedType::Unstaged`
- `--staged` or `--cached` → `UncommittedType::Staged`
- `HEAD` → `UncommittedType::All`
- Other → commit spec (validated for basic safety)

### src/cli/validation.rs (new)

```rust
pub fn validate_diff_has_changes(
    repo_path: &Path,
    spec: &DiffSpec,
) -> Result<(), String>;
```

Checks that the diff actually has changes before starting TUI.

### src/core/diff_source.rs (new)

```rust
pub enum DiffSource { ... }
pub enum UncommittedType { ... }

impl DiffSource {
    pub fn header_text(&self) -> String;
}
```

### src/git/resolve.rs (new)

```rust
pub fn resolve_diff_source(
    repo_path: &Path,
    spec: DiffSpec,
) -> Result<DiffSource, String>;
```

Resolves a `DiffSpec` to a `DiffSource`:
- For commit ranges: resolves refs to commits, counts commits
- For uncommitted: returns directly (no resolution needed)

**Commit range resolution:**
```rust
fn resolve_commit_range(repo_path: &Path, spec: &str) -> Result<CommitRange, String> {
    // Parse spec for .. or ...
    // Resolve refs to commit hashes
    // Count commits in range
    // Return CommitRange { start, end, count }
}
```

### src/git/mod.rs (updated)

Add:
```rust
pub fn fetch_file_changes_for_source(
    repo_path: &Path,
    source: &DiffSource,
) -> Result<Vec<FileChange>, String>;
```

Dispatches to appropriate function based on `DiffSource` variant.

### src/git/uncommitted.rs (new)

```rust
pub fn fetch_uncommitted_file_changes(
    repo_path: &Path,
    uncommitted_type: UncommittedType,
) -> Result<Vec<FileChange>, String>;
```

Uses `git diff` with appropriate flags based on `UncommittedType`.

### src/app/files_view.rs (updated)

**Current:**
```rust
pub range: CommitRange,
```

**After:**
```rust
pub source: DiffSource,
```

The `source` field stores what we're diffing. For commit ranges, we can extract `CommitRange` when needed for diff lookups.

### src/app/mod.rs (updated)

Add:
```rust
impl App {
    pub fn with_diff_source(repo_path: PathBuf, source: DiffSource) -> Self;
}
```

Creates app directly in FilesView without LogView.

### src/main.rs (updated)

```rust
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = env::args().collect();
    let command = cli::parse_args(&args);
    
    let repo_path = resolve_repo_path(&command)?;
    
    match command {
        Command::Log => run_log_view(repo_path),
        Command::Diff(spec) => run_diff_view(repo_path, spec),
    }
}

fn run_diff_view(repo_path: PathBuf, spec: DiffSpec) -> io::Result<()> {
    // Validate before entering TUI
    cli::validate_diff_has_changes(&repo_path, &spec)?;
    
    // Resolve spec to DiffSource
    let source = git::resolve_diff_source(&repo_path, spec)?;
    
    // Fetch file changes
    let files = git::fetch_file_changes_for_source(&repo_path, &source)?;
    
    if files.is_empty() {
        eprintln!("No changes found");
        std::process::exit(1);
    }
    
    // Create app directly in FilesView
    let mut app = App::with_diff_source(repo_path, source);
    
    // ... run TUI loop
}
```

## Interface: CLI → FilesView

```
CLI Args → parse_args() → Command::Diff(DiffSpec)
                                    ↓
                           validate_diff_has_changes()
                                    ↓
                           resolve_diff_source() → DiffSource
                                    ↓
                           fetch_file_changes_for_source()
                                    ↓
                           App::with_diff_source() → FilesView
```

## User Flow

```
$ splice diff main..feature
                    │
                    ▼
┌─────────────────────────────────────────┐
│ Files View                              │
│ main..feature (3 commits)               │  ← header from DiffSource
│                                         │
│ 5 files · +142 -38                      │
│ →├── src/main.rs                        │
│  ├── src/lib.rs                         │
│  └── Cargo.toml                         │
└─────────────────────────────────────────┘
```

```
$ splice diff --staged
                    │
                    ▼
┌─────────────────────────────────────────┐
│ Files View                              │
│ Staged changes                          │  ← header for uncommitted
│                                         │
│ 2 files · +24 -8                        │
│ →├── src/app.rs                         │
│  └── src/lib.rs                         │
└─────────────────────────────────────────┘
```

## Error Handling

All errors happen before entering TUI:

```bash
$ splice diff nonexistent..branch
Error: error resolving start ref "nonexistent": fatal: ambiguous argument

$ splice diff main..main
Error: No changes found

$ splice diff --staged
(No staged changes)
Error: No changes found
```

## Tradeoffs Considered

### Option A: Loading state (like Go)

Go uses a `DirectDiffLoadingState` that shows a loading screen while fetching files.

**Pros:** Consistent async pattern
**Cons:** More complexity, brief flicker for typically fast operations

**Decision:** Skip loading state. Fetch synchronously before entering TUI. If slow, we can add loading state later.

### Option B: Start in LogView, navigate to FilesView

Create the LogView but immediately navigate to FilesView.

**Pros:** Reuses existing initialization
**Cons:** Wastes work fetching commits we don't need

**Decision:** Create App directly in FilesView state.

### Option C: Defer uncommitted changes support

Implement only commit ranges initially, add uncommitted later.

**Pros:** Smaller initial scope
**Cons:** Incomplete feature, harder to add later

**Decision:** Implement full `DiffSource` enum from start. The abstraction cost is minimal and ensures the API is correct.

## Edge Cases

1. **Invalid ref**: Error before TUI, user sees clear message
2. **Empty diff**: Error before TUI, "No changes found"
3. **Merge conflict markers in spec**: Rejected by `is_valid_diff_spec()`
4. **Shell metacharacters**: Rejected for security
5. **Relative refs** (HEAD~5, HEAD^): Supported via git resolution

## Testing Strategy

1. **Unit tests**: `parse_diff_args()`, `is_valid_diff_spec()`, `DiffSource::header_text()`
2. **Integration tests**: `resolve_commit_range()` with `TestRepo`
3. **E2E tests**: CLI invocation with various specs

## Implementation Order

1. Add `DiffSource` and `UncommittedType` to `src/core/diff_source.rs`
2. Add CLI parsing to `src/cli/`
3. Add `resolve_commit_range()` to `src/git/resolve.rs`
4. Add `fetch_uncommitted_file_changes()` to `src/git/uncommitted.rs`
5. Add `fetch_file_changes_for_source()` to `src/git/mod.rs`
6. Update `FilesView` to store `DiffSource`
7. Add `App::with_diff_source()`
8. Update `src/main.rs` to handle diff command
9. Add tests

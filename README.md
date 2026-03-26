> 🚧 **Beta** — Splice is under active development. Things may break or change. Feedback, bug reports, and ideas are very welcome — [open an issue](https://github.com/oberprah/splice/issues).

# Splice

**Git history, the way it should feel.** A terminal-native git log/diff viewer. Lean interface, intuitive navigation, fast keyboard-driven workflow.

<p align="center">
  <img src="assets/demo.gif" alt="Splice demo" width="1200" />
</p>

## Features

- **Commit graph** — visual branch topology with merge lines, so you can actually follow what happened
- **File tree** — expand any commit to see changed files with status indicators and line counts
- **Side-by-side and unified diffs** — toggle between layouts with `v`, with inline highlighting for changed lines
- **Smart navigation** — jump between diffs with `n`/`p`, drill into files with `Enter`, back out with `q` — always one key away
- **Fast** — opens instantly, even on large repos. No config, no setup, no waiting

## Install

```bash
cargo install --path .
```

## Usage

```
splice                        Log view for current repo
splice --all                  All branches
splice diff                   Unstaged changes
splice diff --staged          Staged changes
splice diff main..feature     Compare two refs
```

Run `splice --help` for the full list of options.

## Contributing

Splice is in early development — contributions are welcome! Check out the [open issues](https://github.com/oberprah/splice/issues) or open a new one to get started.

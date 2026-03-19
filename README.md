# Splice

A terminal-native git log/diff viewer. Fast, keyboard-driven, no config needed.

<p align="center">
  <img src="assets/demo.gif" alt="Splice demo" width="800" />
</p>

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

## Development

Setup git hooks:

```bash
git config core.hooksPath .githooks
```

See [CLAUDE.md](CLAUDE.md) for architecture and development guidelines.

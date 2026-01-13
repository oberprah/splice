# Sandbox Environment

Docker-based isolated environment for running Claude Code against this repository.

## Purpose

Provides a controlled testing environment where Claude Code can:
- Execute commands and tests without affecting the host system
- Work with a clean, reproducible setup
- Be safely given dangerously-skip-permissions mode

## Usage

```bash
./sandbox.sh           # Start Claude Code in sandbox
./sandbox.sh shell     # Open bash shell for debugging
./sandbox.sh stop      # Stop containers
./sandbox.sh down      # Remove containers completely
```

## Architecture

- **Dockerfile**: Claude Code container with Go toolchain
- **docker-compose.yml**: Service orchestration
- **template-claude.json**: Initial Claude Code configuration
- **sandbox.sh**: Environment manager (setup, theme sync, lifecycle)

Agent state persists in `../.sandbox-agent-data/` across container restarts.

# Sandbox Environment

Docker-based isolated environment for running Claude Code against this repository.

## Purpose

Provides a controlled testing environment where Claude Code can:
- Execute commands and tests without affecting the host system
- Work with a clean, reproducible setup
- Be safely given dangerously-skip-permissions mode

## Security Model

**Threat: Prompt Injection**
Internet access (web searches, fetches) is required for agent capabilities but creates prompt injection risk from external content.

**Mitigation: No Secrets in Container**
- API keys stay on host, proxied through LiteLLM sidecar
- Git remote operations disabled (no origin push/pull)
- No GitHub CLI or credentials
- Remote operations must be done outside the container

**Isolation: Host Protection**
Container limits prevent host system takeover through resource exhaustion or breakout attempts.

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

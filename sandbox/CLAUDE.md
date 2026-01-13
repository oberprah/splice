# Sandbox Environment

Isolated environment for running Claude Code with command execution and internet access. Allows dangerously-skip-permissions mode while protecting the host system from compromise.

## Security Model

Defense-in-depth strategy against prompt injection and container breakout:

**Layer 1: No Secrets**
- API keys proxied through LiteLLM sidecar (never in container)
- No git remote access (push/pull disabled)
- No GitHub CLI or credentials
- Remote operations done outside container

**Layer 2: Container Isolation**
- Resource limits (CPU, memory, PIDs) prevent exhaustion attacks
- No privileged mode or host access
- Disposable: compromise affects only sandboxed code

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

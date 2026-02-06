# Sandbox Environment

Isolated Docker environment for running Claude Code and Codex CLI with full command execution and internet access.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Security Model

**Threat Model**: Protect host machine from compromised AI agent (prompt injection attacks)

This docker-compose setup provides **basic container isolation** with **accepted security tradeoffs for simplicity**.

### Key Security Features

- Non-root user (UID 1000)
- All capabilities dropped except 4 essential
- Default seccomp profile blocks dangerous syscalls
- Resource limits (4 CPUs, 8GB RAM, 200 PIDs)
- Protected volumes (.github, sandbox are read-only)

### Critical Risks

**🔴 API Key Exposed**: Agent can read `ANTHROPIC_AUTH_TOKEN` from environment
- Mitigation: Set spending limits, monitor usage, use dedicated sandbox key

**🟠 Data Exfiltration**: Agent has internet access and can read all workspace files
- Mitigation: Don't use on repos with secrets/credentials

**🟡 Container Escape**: Single container boundary, requires updated Docker/kernel
- Mitigation: Keep system updated, use dedicated dev machine

### When NOT to use

- Repository contains secrets, API keys, or credentials
- Code is highly sensitive or proprietary
- Compliance requirements exist (SOC2, HIPAA, etc.)

## Setup

**1. Create your docker-compose.yml:**

```bash
cd sandbox
cp docker-compose.yml.template docker-compose.yml
```

**2. Configure docker-compose.yml:**

Edit `sandbox/docker-compose.yml`:

**Optional: Mount Codex config** (uncomment if you want to use Codex CLI):
```yaml
volumes:
  - ~/.codex:/home/agent/.codex
```

**Environment variables** (uncomment what you need):
```yaml
environment:
  - ANTHROPIC_AUTH_TOKEN              # Required for Claude Code
  - ANTHROPIC_BASE_URL                # Optional: corporate proxy
  - OPENAI_API_KEY                    # Optional: for Codex CLI
```

**3. Set environment variables in your shell:**

```bash
# bash/zsh
export ANTHROPIC_AUTH_TOKEN='sk-ant-...'

# fish
set -gx ANTHROPIC_AUTH_TOKEN 'sk-ant-...'
```

To persist across sessions, add to your shell config (`~/.zshrc`, `~/.config/fish/config.fish`).

## Usage

```bash
./sandbox/sandbox.sh           # Start Claude Code (default)
./sandbox/sandbox.sh codex     # Start Codex CLI
./sandbox/sandbox.sh shell     # Open bash shell for debugging
./sandbox/sandbox.sh stop      # Stop sandbox (keeps container)
./sandbox/sandbox.sh down      # Stop and remove sandbox
./sandbox/sandbox.sh status    # Check sandbox status
./sandbox/sandbox.sh logs      # View sandbox logs
```

The first run will build the Docker image (includes Go toolchain, Claude Code CLI, Codex CLI, tmux).

## Architecture

```
Host Machine → Docker Compose → sandbox container
  ├── Claude Code CLI (will configure on first run)
  ├── Codex CLI (optional, mount ~/.codex config)
  ├── Go 1.25.2 toolchain
  ├── tmux (for tape-runner tests)
  └── Internet access
```

## Project Structure

```
sandbox/
├── docker/
│   └── agent/
│       └── Dockerfile                     # Go + Claude + Codex + tmux
├── docker-compose.yml.template            # Base configuration template
├── docker-compose.yml                     # Your personal config (gitignored)
├── sandbox.sh                             # Convenience wrapper script
└── CLAUDE.md                              # This file
```

## What's Included

**Pre-installed tools:**
- Go 1.25.2 toolchain
- Node.js 22.x
- Claude Code CLI (native binary with auto-update)
- OpenAI Codex CLI
- tmux (required for tape-runner tests)

**Volume mounts:**
- `/workspace` - Project root (read/write)
- `/workspace/.github` - CI/CD workflows (read-only)
- `/workspace/sandbox` - Sandbox config (read-only)

**Configuration:**
- Claude Code: Configures on first run (no mount needed)
- Codex CLI: Optional mount of `~/.codex` for your config

## Development Notes

**Testing with tape-runner:**
- The `./run-tape` script works inside the sandbox
- Run with `./sandbox/sandbox.sh shell` then `./run-tape --help`

**Agent state:**
- State is stored inside the container (not persisted to host)
- Use `./sandbox/sandbox.sh down` to destroy all state
- State persists between stops/starts as long as container exists

**Rebuilding after changes:**
```bash
docker compose -f sandbox/docker-compose.yml build
./sandbox/sandbox.sh down && ./sandbox/sandbox.sh up
```

# Sandbox Environment

Isolated Docker environment for running AI coding agents (Claude Code, Codex CLI, OpenCode) with full command execution and internet access.

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

**Optional: Mount AI agent configs** (uncomment what you need):
```yaml
volumes:
  # Codex CLI
  - ~/.codex:/home/agent/.codex

  # OpenCode
  - ~/.local/share/opencode:/home/agent/.local/share/opencode
  - ~/.config/opencode:/home/agent/.config/opencode
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
./sandbox/sandbox.sh opencode  # Start OpenCode server
./sandbox/sandbox.sh shell     # Open bash shell
./sandbox/sandbox.sh stop      # Stop sandbox (keeps container)
./sandbox/sandbox.sh down      # Stop and remove sandbox
./sandbox/sandbox.sh status    # Check sandbox status
./sandbox/sandbox.sh logs      # View sandbox logs
```

**Codex CLI Usage:**

Codex CLI runs with `--yolo` (disables internal sandboxing since Docker provides isolation) and `-c shell_environment_policy.inherit=all` to pass through all environment variables. The mise shims PATH is added to `~/.profile` so login shells can access tools from `.mise.toml`.

**OpenCode Usage:**

OpenCode's TUI doesn't work directly inside Docker containers ([known issue](https://github.com/anomalyco/opencode/issues/12439)). We use server mode instead:

```bash
./sandbox/sandbox.sh opencode
```

This automatically:
- Starts OpenCode server in the container (if not already running)
- Connects your host terminal to it

You get OpenCode's full TUI experience on your host terminal, while the agent executes commands inside the sandboxed container. Run the command multiple times - it's safe and will reuse the existing server.

The first run will build the Docker image (includes Go toolchain, Claude Code CLI, Codex CLI, OpenCode, tmux).

## Architecture

```
Host Machine → Docker Compose → sandbox container
  ├── mise (language/tool version manager)
  │   └── Shims added to PATH via ~/.profile
  ├── Claude Code CLI (will configure on first run)
  ├── Codex CLI (optional, mount ~/.codex config)
  │   └── Runs with shell_environment_policy.inherit=all
  ├── OpenCode (optional, mount ~/.local/share/opencode and ~/.config/opencode)
  ├── Node.js 24.x (global via mise)
  ├── Go 1.25.2 (from .mise.toml)
  ├── tmux (for tape-runner tests)
  └── Internet access
```

### Codex CLI Environment Configuration

Codex CLI uses a restricted PATH by default for security. Since we're in a sandboxed container, we disable this restriction with `shell_environment_policy.inherit=all` to access mise-managed tools.

**Changing language toolchains:**

Edit `.mise.toml` in the repo root to add/change languages:

```toml
[tools]
go = "1.25.2"      # Or remove for Rust migration
rust = "latest"    # Add when migrating to Rust
```

Then rebuild: `docker compose -f sandbox/docker-compose.yml build`

## Project Structure

```
.mise.toml                                 # Language toolchain config (Go, Rust, etc.)
sandbox/
├── docker/
│   └── agent/
│       └── Dockerfile                     # Debian + mise + AI CLIs
├── docker-compose.yml.template            # Base configuration template
├── docker-compose.yml                     # Your personal config (gitignored)
├── sandbox.sh                             # Convenience wrapper script
└── CLAUDE.md                              # This file
```

## What's Included

**Pre-installed tools:**
- [mise](https://mise.jdx.dev/) - Universal version manager for languages and tools
- Node.js 24.x (via mise, for AI agent CLIs)
- Claude Code CLI (native binary with auto-update)
- OpenAI Codex CLI
- OpenCode
- tmux (required for tape-runner tests)

**Language toolchains** (managed via `.mise.toml`):
- Go 1.25.2 (configured in `.mise.toml` at repo root)
- Add/change languages by editing `.mise.toml` - no Dockerfile changes needed

**Volume mounts:**
- `/workspace` - Project root (read/write)
- `/workspace/.github` - CI/CD workflows (read-only)
- `/workspace/sandbox` - Sandbox config (read-only)

**Configuration:**
- Claude Code: Configures on first run (no mount needed)
- Codex CLI: Optional mount of `~/.codex` for your config
- OpenCode: Optional mount of `~/.local/share/opencode` and `~/.config/opencode`

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

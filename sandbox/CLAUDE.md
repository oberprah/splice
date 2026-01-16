# Sandbox Environment

Isolated Kubernetes environment for running Claude Code with command execution and internet access. Uses Kind (Kubernetes in Docker) for defense-in-depth security.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Security Model

Defense-in-depth strategy against prompt injection and container breakout:

**Layer 1: Network Isolation (NetworkPolicy)**
- Agent can only reach the API proxy
- All other egress blocked (no internet, no git push)

**Layer 2: API Proxy (Envoy)**
- API keys injected by proxy (agent never sees real tokens)
- All requests logged for audit

**Layer 3: Container Hardening**
- Non-root user (UID 1000)
- No privilege escalation
- All capabilities dropped
- Seccomp profile restricts syscalls

**Layer 4: Kind Isolation**
- Kubernetes runs inside Docker container
- Escape requires breaking out of pod AND kind container

## Usage

```bash
export AGENT_SANDBOX_TOKEN="your-api-key"

./sandbox/scripts/sandbox.sh           # Start Claude Code in sandbox
./sandbox/scripts/sandbox.sh codex     # Start Codex CLI in sandbox
./sandbox/scripts/sandbox.sh shell     # Open bash shell for debugging
./sandbox/scripts/sandbox.sh stop      # Stop pods (fast restart later)
./sandbox/scripts/sandbox.sh down      # Remove cluster completely
./sandbox/scripts/sandbox.sh status    # Check sandbox status
```

## Architecture

```
Host Machine
  |
  v
Kind (Docker container running Kubernetes)
  |
  v
agent-env namespace
  +-- claude-agent pod (Claude Code + Codex CLI + Go toolchain, no internet)
  +-- api-proxy pod (Envoy, injects credentials, routes to Anthropic/OpenAI APIs)
```

## Project Structure

```
sandbox/
+-- docker/
|   +-- agent/
|   |   +-- Dockerfile        # Go toolchain + Claude Code + Codex CLI
|   |   +-- claude.json       # Pre-configured Claude settings
|   |   +-- codex-config.toml # Pre-configured Codex settings
|   +-- proxy/
|       +-- Dockerfile        # Envoy proxy
|       +-- envoy.yaml        # API routing and credential injection
+-- k8s/
|   +-- namespace.yaml
|   +-- agent-pod.yaml
|   +-- proxy-pod.yaml
|   +-- proxy-service.yaml
|   +-- network-policy.yaml
+-- scripts/
|   +-- sandbox.sh          # Start, stop, shell commands
+-- kind-config.yaml        # Kind cluster configuration
+-- SECURITY.md             # Threat model and isolation layers
```

Agent state persists in the project directory across container restarts.

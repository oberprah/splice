#!/bin/bash
# Sandbox environment manager using Docker Compose

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="$(basename "$(dirname "$SCRIPT_DIR")")-sandbox"

# Generate a deterministic port (4096-5095) based on directory name
# This ensures each repo copy gets a unique, consistent port
DIR_HASH=$(echo "$PROJECT_NAME" | cksum | cut -d' ' -f1)
OPENCODE_PORT=$((4096 + (DIR_HASH % 1000)))

# Check if docker-compose.yml exists
if [ ! -f "$SCRIPT_DIR/docker-compose.yml" ]; then
    cat >&2 <<EOF
Error: docker-compose.yml not found

Please create your personal docker-compose.yml configuration:

  cd sandbox
  cp docker-compose.yml.template docker-compose.yml

Then edit docker-compose.yml to configure:
  - Config file mounts (volumes section)
  - Environment variables (environment section)

See sandbox/CLAUDE.md for detailed setup instructions.
EOF
    exit 1
fi

# Note: Environment variable validation removed - now configured per-user in docker-compose.yml

# Export port for docker-compose.yml
export OPENCODE_PORT

# Check if container is running
is_running() {
    docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" ps -q agent | grep -q .
}

# Ensure container is running
ensure_running() {
    if ! is_running; then
        echo "-> Starting sandbox..."
        docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" up -d
        echo ""
        echo "Sandbox ready (OpenCode port: $OPENCODE_PORT)"
        echo ""
    fi
}

# Main command dispatcher
case "${1:-claude}" in
    up)
        echo "-> Starting sandbox..."
        docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" up -d
        echo ""
        echo "Sandbox ready (OpenCode port: $OPENCODE_PORT)"
        ;;

    claude|"")
        ensure_running
        docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" exec agent claude --dangerously-skip-permissions
        ;;

    codex)
        ensure_running
        docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" exec agent codex --yolo -c shell_environment_policy.inherit=all
        ;;

    opencode)
        ensure_running

        # Check if OpenCode server is already running
        if docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" exec -T agent pgrep -f "opencode serve" > /dev/null 2>&1; then
            echo "-> OpenCode server already running on port $OPENCODE_PORT"
        else
            echo "-> Starting OpenCode server on port $OPENCODE_PORT..."
            docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" exec -d agent opencode serve --hostname 0.0.0.0 --port 4096
            echo "-> Waiting for server to start..."
            sleep 3
        fi

        echo "-> Connecting to OpenCode..."
        opencode attach http://localhost:$OPENCODE_PORT
        ;;

    shell)
        ensure_running
        docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" exec agent /bin/bash
        ;;

    stop)
        echo "-> Stopping sandbox..."
        docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" stop
        echo "Stopped"
        ;;

    down)
        echo "-> Stopping and removing sandbox..."
        docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" down
        echo "Removed"
        ;;

    logs)
        docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" logs -f agent
        ;;

    status)
        if is_running; then
            echo "Sandbox: running (OpenCode port: $OPENCODE_PORT)"
            docker compose -p "$PROJECT_NAME" -f "$SCRIPT_DIR/docker-compose.yml" ps
        else
            echo "Sandbox: not running"
        fi
        ;;

    *)
        cat <<EOF
Usage: $0 [command]

Commands:
  up        Start sandbox container
  claude    Start Claude Code in sandbox (default)
  codex     Start OpenAI Codex CLI in sandbox
  opencode  Start OpenCode (auto-connects to sandbox)
  shell     Open bash shell in sandbox
  stop      Stop sandbox (keeps container)
  down      Stop and remove sandbox
  logs      View sandbox logs
  status    Show sandbox status

Examples:
  $0              # Start Claude Code
  $0 codex        # Start Codex CLI
  $0 opencode     # Start OpenCode and connect
  $0 shell        # Debug with bash
  $0 stop         # Stop container
  $0 down         # Stop and remove
EOF
        exit 1
        ;;
esac

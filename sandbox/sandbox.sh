#!/bin/bash
# Sandbox environment manager using Docker Compose

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

# Check if container is running
is_running() {
    docker compose -f "$SCRIPT_DIR/docker-compose.yml" ps -q sandbox | grep -q .
}

# Ensure container is running
ensure_running() {
    if ! is_running; then
        echo "-> Starting sandbox..."
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" up -d
        echo ""
        echo "Sandbox ready"
        echo ""
    fi
}

# Main command dispatcher
case "${1:-claude}" in
    up)
        echo "-> Starting sandbox..."
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" up -d
        echo ""
        echo "Sandbox ready"
        ;;

    claude|"")
        ensure_running
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" exec sandbox claude --dangerously-skip-permissions
        ;;

    codex)
        ensure_running
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" exec sandbox codex --yolo
        ;;

    shell)
        ensure_running
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" exec sandbox /bin/bash
        ;;

    stop)
        echo "-> Stopping sandbox..."
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" stop
        echo "Stopped"
        ;;

    down)
        echo "-> Stopping and removing sandbox..."
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" down
        echo "Removed"
        ;;

    logs)
        docker compose -f "$SCRIPT_DIR/docker-compose.yml" logs -f sandbox
        ;;

    status)
        if is_running; then
            echo "Sandbox: running"
            docker compose -f "$SCRIPT_DIR/docker-compose.yml" ps
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
  shell     Open bash shell in sandbox
  stop      Stop sandbox (keeps container)
  down      Stop and remove sandbox
  logs      View sandbox logs
  status    Show sandbox status

Examples:
  $0              # Start Claude Code
  $0 codex        # Start Codex CLI
  $0 shell        # Debug with bash
  $0 stop         # Stop container
  $0 down         # Stop and remove
EOF
        exit 1
        ;;
esac

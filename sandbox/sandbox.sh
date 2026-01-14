#!/bin/bash
# Sandbox environment manager

set -e

cd "$(dirname "$0")"

REPO_ROOT="$(cd .. && pwd)"
AGENT_DATA_DIR="$REPO_ROOT/.sandbox-agent-data"
AGENT_CONFIG="$AGENT_DATA_DIR/claude.json"
AGENT_DIR="$AGENT_DATA_DIR/claude-dir"
TEMPLATE_CONFIG="./template-claude.json"

# Initialize agent config directory
init_agent_config() {
    if [ ! -d "$AGENT_DATA_DIR" ]; then
        echo "→ Creating agent data directory..."
        mkdir -p "$AGENT_DATA_DIR"
    fi

    if [ ! -f "$AGENT_CONFIG" ]; then
        echo "→ Initializing Claude config from template..."
        cp "$TEMPLATE_CONFIG" "$AGENT_CONFIG"
    fi

    if [ ! -d "$AGENT_DIR" ]; then
        echo "→ Creating Claude settings directory..."
        mkdir -p "$AGENT_DIR"
    fi
}

# Sync theme from host Claude config
sync_theme() {
    local HOST_CONFIG="$HOME/.claude.json"

    if [ -f "$HOST_CONFIG" ]; then
        # Extract theme from host config using grep/sed (portable, no jq required)
        local THEME=$(grep -o '"theme"[[:space:]]*:[[:space:]]*"[^"]*"' "$HOST_CONFIG" | sed 's/.*"\([^"]*\)".*/\1/')

        if [ -n "$THEME" ]; then
            echo "→ Syncing theme: $THEME"
            # Update theme in agent config
            sed -i.bak "s/\"theme\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"theme\": \"$THEME\"/" "$AGENT_CONFIG"
            rm -f "$AGENT_CONFIG.bak"
        fi
    fi
}

# Build images if needed
ensure_built() {
    if ! docker compose images sandbox | grep -q sandbox; then
        echo "→ Building images (first run)..."
        docker compose build
    fi
}

# Ensure containers are running
ensure_running() {
    if ! docker compose ps sandbox | grep -q "Up"; then
        echo "→ Starting sandbox environment..."
        docker compose up -d

        # Wait for litellm health check
        echo "→ Waiting for services to be healthy..."
        docker compose exec sandbox echo "Ready" > /dev/null 2>&1 || sleep 2
    fi
}

# Main command dispatcher
case "${1:-claude}" in
    claude|"")
        ensure_built
        init_agent_config
        sync_theme
        ensure_running
        docker compose exec sandbox claude --dangerously-skip-permissions
        ;;

    shell)
        ensure_built
        init_agent_config
        ensure_running
        docker compose exec sandbox /bin/bash
        ;;

    stop)
        echo "→ Stopping containers..."
        docker compose stop
        echo "✓ Stopped"
        ;;

    down)
        echo "→ Stopping and removing containers..."
        docker compose down
        echo "✓ Removed"
        ;;

    *)
        cat <<EOF
Usage: $0 [command]

Commands:
  (none)    Start Claude Code (default)
  shell     Open bash shell in sandbox
  stop      Stop containers (keeps them for fast restart)
  down      Stop and remove containers

Examples:
  $0           # Start Claude Code
  $0 shell     # Open shell for debugging
  $0 stop      # Stop containers
  $0 down      # Remove containers completely
EOF
        exit 1
        ;;
esac

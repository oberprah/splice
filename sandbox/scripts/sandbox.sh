#!/bin/bash
# Sandbox environment manager using Kind (Kubernetes in Docker)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SANDBOX_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$SANDBOX_DIR")"
CLUSTER_NAME="sandbox-$(basename "$PROJECT_ROOT")"

# Proxy configuration with defaults
: "${PROXY_HOST:=taia.tngtech.com}"
: "${PROXY_ANTHROPIC_PREFIX:=/proxy/anthropic/}"
: "${PROXY_OPENAI_PREFIX:=/proxy/openai/}"
export PROXY_HOST PROXY_ANTHROPIC_PREFIX PROXY_OPENAI_PREFIX

# Require API token from environment
require_token() {
    : "${AGENT_SANDBOX_TOKEN:?Set AGENT_SANDBOX_TOKEN environment variable}"
}

# Check if cluster exists
cluster_exists() {
    kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"
}

# Check if agent pod is running
agent_running() {
    kubectl get pod claude-agent -n agent-env --no-headers 2>/dev/null | grep -q "Running"
}

# Create cluster and deploy everything
create_cluster() {
    echo "Project root: $PROJECT_ROOT"
    echo ""

    # Generate kind config with project path
    echo "-> Generating kind config..."
    export PROJECT_ROOT
    envsubst < "$SANDBOX_DIR/kind-config.yaml" > "$SANDBOX_DIR/kind-config-resolved.yaml"

    echo "-> Creating kind cluster with project mount..."
    kind create cluster --name "$CLUSTER_NAME" --config "$SANDBOX_DIR/kind-config-resolved.yaml"

    # Cleanup generated config
    rm -f "$SANDBOX_DIR/kind-config-resolved.yaml"

    build_and_deploy
}

# Build images and deploy to existing cluster
build_and_deploy() {
    require_token

    echo "-> Generating proxy config..."
    envsubst < "$SANDBOX_DIR/docker/proxy/envoy.yaml.template" > "$SANDBOX_DIR/docker/proxy/envoy.yaml"

    echo "-> Building images..."
    docker build -q -t claude-agent:local "$SANDBOX_DIR/docker/agent/"
    docker build -q -t api-proxy:local "$SANDBOX_DIR/docker/proxy/"

    echo "-> Loading images into kind..."
    kind load docker-image claude-agent:local --name "$CLUSTER_NAME"
    kind load docker-image api-proxy:local --name "$CLUSTER_NAME"

    echo "-> Applying Kubernetes manifests..."
    kubectl apply -f "$SANDBOX_DIR/k8s/namespace.yaml"

    # Create secret from env var (never stored in git)
    kubectl create secret generic api-credentials \
        --from-literal=AGENT_SANDBOX_TOKEN="$AGENT_SANDBOX_TOKEN" \
        --namespace=agent-env \
        --dry-run=client -o yaml | kubectl apply -f -

    kubectl apply -f "$SANDBOX_DIR/k8s/network-policy.yaml"

    echo "-> Starting proxy..."
    kubectl apply -f "$SANDBOX_DIR/k8s/proxy-pod.yaml"
    kubectl apply -f "$SANDBOX_DIR/k8s/proxy-service.yaml"
    kubectl wait --for=condition=Ready pod/api-proxy -n agent-env --timeout=120s

    echo "-> Starting agent..."
    kubectl apply -f "$SANDBOX_DIR/k8s/agent-pod.yaml"
    kubectl wait --for=condition=Ready pod/claude-agent -n agent-env --timeout=120s

    echo ""
    echo "Sandbox ready"
}

# Ensure cluster and pods are running
ensure_running() {
    if ! cluster_exists; then
        create_cluster
    elif ! agent_running; then
        echo "-> Cluster exists, restarting pods..."
        build_and_deploy
    fi
}

# Main command dispatcher
case "${1:-claude}" in
    claude|"")
        ensure_running
        echo ""
        kubectl exec -it -n agent-env claude-agent -- claude --dangerously-skip-permissions
        ;;

    codex)
        ensure_running
        echo ""
        kubectl exec -it -n agent-env claude-agent -- codex --full-auto
        ;;

    shell)
        ensure_running
        echo ""
        kubectl exec -it -n agent-env claude-agent -- /bin/bash
        ;;

    stop)
        echo "-> Stopping pods (keeping cluster for fast restart)..."
        kubectl delete pod claude-agent api-proxy -n agent-env --ignore-not-found --grace-period=1
        echo "Stopped"
        ;;

    down)
        echo "-> Deleting cluster..."
        kind delete cluster --name "$CLUSTER_NAME"
        echo "Removed"
        ;;

    status)
        if cluster_exists; then
            echo "Cluster: running"
            kubectl get pods -n agent-env 2>/dev/null || echo "Namespace not found"
        else
            echo "Cluster: not running"
        fi
        ;;

    *)
        cat <<EOF
Usage: $0 [command]

Commands:
  claude    Start Claude Code in sandbox (default)
  codex     Start OpenAI Codex CLI in sandbox
  shell     Open bash shell in sandbox
  stop      Stop pods (keeps cluster for fast restart)
  down      Delete cluster completely
  status    Show sandbox status

Examples:
  $0              # Start Claude Code
  $0 codex        # Start Codex CLI
  $0 shell        # Debug with bash
  $0 stop         # Quick stop, fast restart later
  $0 down         # Full cleanup
EOF
        exit 1
        ;;
esac

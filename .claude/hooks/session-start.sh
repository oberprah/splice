#!/bin/bash
# SessionStart hook for Claude Code
# Sets up development environment with Go, git hooks, and required tools

set -euo pipefail

# Get the directory where this script is located
readonly HOOKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ============================================================================
# Git Hooks Configuration (runs locally and in Claude Code web)
# ============================================================================

cd "${CLAUDE_PROJECT_DIR}"
git config core.hooksPath .githooks 2>/dev/null || true

# ============================================================================
# Claude Code Web-Only Setup
# ============================================================================

# Early exit for local development - remaining setup only needed in Claude Code web
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

# Prevent Go from auto-downloading toolchains before we check versions
export GOTOOLCHAIN=local

# Check if Go is already installed at correct version
required_version=$(awk '/^go [0-9]/ {print $2}' "${CLAUDE_PROJECT_DIR}/go.mod")
current_version=""
if command -v go &> /dev/null; then
  current_version=$(go version | awk '{print $3}' | sed 's/go//')
fi

# Setup Go if needed
if [[ "$current_version" != "$required_version" ]]; then
  source "${HOOKS_DIR}/setup-go.sh"
  setup_go > /dev/null 2>&1
  status="Go ${required_version} ✓"
else
  status="Go ${required_version} ✓"
fi

echo "🔧 Environment ready: ${status}  |  Git hooks ✓  |  Tools: on-demand"

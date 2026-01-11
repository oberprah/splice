#!/bin/bash
# SessionStart hook for Claude Code
# Sets up development environment with Go, git hooks, and required tools

set -euo pipefail

# Get the directory where this script is located
readonly HOOKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ============================================================================
# Git Hooks Configuration (runs locally and in Claude Code web)
# ============================================================================

echo "🔧 Configuring git hooks..."
cd "${CLAUDE_PROJECT_DIR}"
git config core.hooksPath .githooks
echo "✅ Git hooks configured"
echo ""

# ============================================================================
# Claude Code Web-Only Setup
# ============================================================================

# Early exit for local development - remaining setup only needed in Claude Code web
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

echo "🔧 Setting up Claude Code web environment..."
echo ""

# Prevent Go from auto-downloading toolchains before we check versions
export GOTOOLCHAIN=local

source "${HOOKS_DIR}/setup-go.sh"
setup_go
echo ""

source "${HOOKS_DIR}/setup-devtools.sh"
setup_devtools
echo ""

echo "✅ Environment setup complete"

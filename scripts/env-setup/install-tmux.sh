#!/bin/bash
# Install tmux based on the current environment

set -euo pipefail

install_tmux() {
  # Check if already installed
  if command -v tmux &> /dev/null; then
    echo "✅ tmux is already installed"
    return 0
  fi

  # Claude Code web (Ubuntu container)
  if [ "${CLAUDE_CODE_REMOTE:-}" = "true" ]; then
    echo "📦 Installing tmux (Claude Code web environment)..."
    apt-get update -qq
    apt-get install -y tmux
    echo "✅ tmux installed successfully"
    return 0
  fi

  # macOS (don't auto-install, just inform)
  if [ "$(uname)" = "Darwin" ]; then
    echo "ℹ️  tmux is not installed"
    echo "   Install with: brew install tmux"
    exit 1
  fi

  # Other Linux distributions
  if command -v apt-get &> /dev/null; then
    echo "📦 Installing tmux (Debian/Ubuntu)..."
    sudo apt-get update -qq
    sudo apt-get install -y tmux
    echo "✅ tmux installed successfully"
    return 0
  fi

  # Unsupported environment
  echo "❌ Unsupported environment for auto-install"
  echo "   Please install tmux manually"
  exit 1
}

# Execute installation when run directly
install_tmux

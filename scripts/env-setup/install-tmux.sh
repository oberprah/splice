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

  # All other environments - manual installation required
  echo "ℹ️  tmux is not installed"
  echo "   Please install tmux manually:"
  echo "   - macOS: brew install tmux"
  echo "   - Debian/Ubuntu: sudo apt-get install tmux"
  echo "   - Other: See https://github.com/tmux/tmux/wiki/Installing"
  exit 1
}

# Execute installation when run directly
install_tmux

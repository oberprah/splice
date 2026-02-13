#!/bin/bash
# Install tmux based on the current environment

set -euo pipefail

install_tmux() {
  # Check if already installed
  if command -v tmux &> /dev/null; then
    return 0
  fi

  # Claude Code web (Ubuntu container)
  if [ "${CLAUDE_CODE_REMOTE:-}" = "true" ]; then
    echo "Installing tmux..."
    apt-get update -qq 2>&1 | grep -v "^[WE]:" || true
    apt-get install -y -qq tmux > /dev/null 2>&1
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

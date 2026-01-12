#!/bin/bash
# Install freeze and dependencies based on the current environment

set -euo pipefail

install_freeze() {
  # Check if already installed
  if command -v freeze &> /dev/null; then
    echo "✅ freeze is already installed"
    return 0
  fi

  # Claude Code web (Ubuntu container)
  if [ "${CLAUDE_CODE_REMOTE:-}" = "true" ]; then
    echo "📦 Installing freeze and dependencies (Claude Code web environment)..."

    # Install librsvg2-bin (required for PNG rendering)
    if ! command -v rsvg-convert &> /dev/null; then
      echo "   Installing librsvg2-bin..."
      apt-get update -qq 2>&1 | grep -v "^[WE]:" || true
      DEBIAN_FRONTEND=noninteractive apt-get install -y -qq librsvg2-bin > /dev/null 2>&1
    fi

    # Install freeze using Go
    echo "   Installing freeze..."
    go install github.com/charmbracelet/freeze@latest

    # Add Go bin directory to PATH for current session
    if [[ -n "${CLAUDE_ENV_FILE:-}" ]]; then
      GO_BIN_DIR="${HOME}/go/bin"
      if [[ -d "$GO_BIN_DIR" ]]; then
        echo "export PATH=\"${GO_BIN_DIR}:\$PATH\"" >> "$CLAUDE_ENV_FILE"
      fi
    fi

    echo "✅ freeze installed successfully"
    return 0
  fi

  # macOS (don't auto-install, just inform)
  if [ "$(uname)" = "Darwin" ]; then
    echo "ℹ️  freeze is not installed"
    echo "   Install with: brew install charmbracelet/tap/freeze"
    exit 1
  fi

  # Other Linux distributions
  if command -v go &> /dev/null; then
    echo "📦 Installing freeze..."

    # Install librsvg2-bin if on Debian/Ubuntu
    if command -v apt-get &> /dev/null && ! command -v rsvg-convert &> /dev/null; then
      echo "   Installing librsvg2-bin..."
      sudo apt-get update -qq 2>&1 | grep -v "^[WE]:" || true
      sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq librsvg2-bin > /dev/null 2>&1
    fi

    go install github.com/charmbracelet/freeze@latest
    echo "✅ freeze installed successfully"
    echo "   Make sure ${HOME}/go/bin is in your PATH"
    return 0
  fi

  # Unsupported environment
  echo "❌ Unsupported environment for auto-install"
  echo "   Please install freeze manually: https://github.com/charmbracelet/freeze"
  exit 1
}

# Execute installation when run directly
install_freeze

#!/bin/bash
# Install freeze and dependencies based on the current environment

set -euo pipefail

install_freeze() {
  # Check if already installed
  if command -v freeze &> /dev/null; then
    return 0
  fi

  # Claude Code web (Ubuntu container)
  if [ "${CLAUDE_CODE_REMOTE:-}" = "true" ]; then
    echo "Installing freeze..."
    # Install librsvg2-bin (required for PNG rendering)
    if ! command -v rsvg-convert &> /dev/null; then
      apt-get update -qq 2>&1 | grep -v "^[WE]:" || true
      DEBIAN_FRONTEND=noninteractive apt-get install -y -qq librsvg2-bin > /dev/null 2>&1
    fi

    # Install freeze using Go
    go install github.com/charmbracelet/freeze@latest 2>&1 | grep -v "^go: downloading" || true

    # Add Go bin directory to PATH for current session
    if [[ -n "${CLAUDE_ENV_FILE:-}" ]]; then
      GO_BIN_DIR="${HOME}/go/bin"
      if [[ -d "$GO_BIN_DIR" ]]; then
        echo "export PATH=\"${GO_BIN_DIR}:\$PATH\"" >> "$CLAUDE_ENV_FILE"
      fi
    fi
    return 0
  fi

  # All other environments - manual installation required
  echo "ℹ️  freeze is not installed"
  echo "   Please install freeze manually:"
  echo "   - macOS: brew install charmbracelet/tap/freeze"
  echo "   - Other: go install github.com/charmbracelet/freeze@latest"
  echo "   - See: https://github.com/charmbracelet/freeze"
  exit 1
}

# Execute installation when run directly
install_freeze

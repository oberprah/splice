#!/bin/bash
# Development tools installation (tmux, freeze) for Claude Code web sessions

set -euo pipefail

# ============================================================================
# Configuration
# ============================================================================

# These variables are used by freeze installation
# GO_INSTALL_DIR should be set by setup-go.sh when sourced
: "${GO_INSTALL_DIR:=/usr/local/go}"
: "${GO_BIN_DIR:=${HOME}/go/bin}"

# ============================================================================
# tmux Setup
# ============================================================================

is_tmux_installed() {
  command -v tmux &> /dev/null
}

install_tmux() {
  echo "📦 Installing tmux..."
  apt-get update -qq
  apt-get install -y tmux
  echo "✅ tmux installed successfully"
}

setup_tmux() {
  if is_tmux_installed; then
    echo "✅ tmux is already installed"
  else
    install_tmux
  fi
}

# ============================================================================
# freeze Setup
# ============================================================================

is_freeze_installed() {
  command -v freeze &> /dev/null
}

install_freeze() {
  echo "📦 Installing freeze..."
  # Use the Go we just set up to install freeze
  "${GO_INSTALL_DIR}/bin/go" install github.com/charmbracelet/freeze@latest

  # Add Go bin directory to PATH if not already there
  if [[ -n "${CLAUDE_ENV_FILE:-}" ]] && [[ -d "$GO_BIN_DIR" ]]; then
    echo "export PATH=\"${GO_BIN_DIR}:\$PATH\"" >> "$CLAUDE_ENV_FILE"
  fi

  echo "✅ freeze installed successfully"
}

setup_freeze() {
  if is_freeze_installed; then
    echo "✅ freeze is already installed"
  else
    install_freeze
  fi
}

# ============================================================================
# Main Setup Function
# ============================================================================

setup_devtools() {
  echo "📦 Setting up development tools..."
  setup_tmux
  setup_freeze
}

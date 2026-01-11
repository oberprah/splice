#!/bin/bash
set -euo pipefail

# Only run in Claude Code on the web
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

# Prevent Go from auto-downloading toolchains before we check versions
export GOTOOLCHAIN=local

# ============================================================================
# Configuration
# ============================================================================

readonly GO_INSTALL_DIR="/usr/local/go"
readonly GO_CACHE_DIR="/tmp"
readonly GO_BIN_DIR="${HOME}/go/bin"

# ============================================================================
# Helper Functions
# ============================================================================

get_required_go_version() {
  # Extract Go version from go.mod (e.g., "go 1.25.2" → "1.25.2")
  awk '/^go [0-9]/ {print $2}' "${CLAUDE_PROJECT_DIR}/go.mod"
}

get_current_go_version() {
  if ! command -v go &> /dev/null; then
    echo ""
    return
  fi
  go version | awk '{print $3}' | sed 's/go//'
}

is_correct_version_installed() {
  local required="$1"
  local current
  current=$(get_current_go_version)

  [[ "$current" == "$required" ]]
}

download_go() {
  local version="$1"
  local archive="go${version}.linux-amd64.tar.gz"
  local url="https://go.dev/dl/${archive}"
  local cache_path="${GO_CACHE_DIR}/${archive}"

  if [[ -f "$cache_path" ]]; then
    echo "✅ Using cached Go archive"
    return 0
  fi

  echo "⬇️  Downloading Go ${version}..."
  curl -L -o "$cache_path" "$url"
}

install_go() {
  local version="$1"
  local archive="go${version}.linux-amd64.tar.gz"
  local cache_path="${GO_CACHE_DIR}/${archive}"

  echo "📦 Installing Go to ${GO_INSTALL_DIR}..."
  rm -rf "$GO_INSTALL_DIR"
  tar -C /usr/local -xzf "$cache_path"
}

verify_installation() {
  local version="$1"

  if ! "${GO_INSTALL_DIR}/bin/go" version | grep -q "$version"; then
    echo "❌ Go installation failed"
    exit 1
  fi

  echo "✅ Go ${version} installed successfully"
}

configure_environment() {
  # Add Go to PATH and configure toolchain for the session
  if [[ -n "${CLAUDE_ENV_FILE:-}" ]]; then
    echo "export PATH=\"${GO_INSTALL_DIR}/bin:\$PATH\"" >> "$CLAUDE_ENV_FILE"
    echo "export GOTOOLCHAIN=local" >> "$CLAUDE_ENV_FILE"
    echo "🔧 Go added to session PATH with local toolchain"
  fi
}

setup_git_hooks() {
  echo "🔧 Configuring git hooks..."
  cd "${CLAUDE_PROJECT_DIR}"
  git config core.hooksPath .githooks
  echo "✅ Git hooks configured"
}

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
# Main
# ============================================================================

main() {
  echo "🔧 Setting up development environment..."
  echo ""

  # ============================================================================
  # Go Installation
  # ============================================================================

  echo "📦 Setting up Go..."

  local required_version
  required_version=$(get_required_go_version)

  # Check if already installed
  if is_correct_version_installed "$required_version"; then
    echo "✅ Go ${required_version} is already installed"
    configure_environment
  else
    # Show upgrade/install message
    local current_version
    current_version=$(get_current_go_version)
    if [[ -n "$current_version" ]]; then
      echo "📦 Found Go ${current_version}, upgrading to ${required_version}..."
    else
      echo "📦 Installing Go ${required_version}..."
    fi

    # Download, install, and verify
    download_go "$required_version"
    install_go "$required_version"
    verify_installation "$required_version"
    configure_environment
  fi

  echo ""

  # ============================================================================
  # Git Hooks
  # ============================================================================

  setup_git_hooks
  echo ""

  # ============================================================================
  # Development Tools
  # ============================================================================

  echo "📦 Setting up development tools..."
  setup_tmux
  setup_freeze

  echo ""
  echo "✅ Environment setup complete"
}

main

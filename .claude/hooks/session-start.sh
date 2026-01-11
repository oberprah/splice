#!/bin/bash
set -euo pipefail

# Only run in Claude Code on the web
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

# ============================================================================
# Configuration
# ============================================================================

readonly GO_INSTALL_DIR="/usr/local/go"
readonly GO_CACHE_DIR="/tmp"

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
  # Add Go to PATH for the session
  if [[ -n "${CLAUDE_ENV_FILE:-}" ]]; then
    echo "export PATH=\"${GO_INSTALL_DIR}/bin:\$PATH\"" >> "$CLAUDE_ENV_FILE"
    echo "🔧 Go added to session PATH"
  fi
}

# ============================================================================
# Main
# ============================================================================

main() {
  echo "🔧 Setting up Go environment..."

  local required_version
  required_version=$(get_required_go_version)

  # Check if already installed
  if is_correct_version_installed "$required_version"; then
    echo "✅ Go ${required_version} is already installed"
    configure_environment
    exit 0
  fi

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

  echo "✅ Environment setup complete"
}

main

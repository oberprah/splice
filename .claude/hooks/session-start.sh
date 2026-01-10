#!/bin/bash
set -euo pipefail

# Only run in Claude Code on the web
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

echo "🔧 Setting up Go environment..."

# Required Go version
REQUIRED_GO_VERSION="1.25.2"
GO_ARCHIVE="go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/${GO_ARCHIVE}"

# Check if correct Go version is already installed
if command -v go &> /dev/null; then
  CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
  if [ "$CURRENT_VERSION" = "$REQUIRED_GO_VERSION" ]; then
    echo "✅ Go $REQUIRED_GO_VERSION is already installed"
    exit 0
  fi
  echo "📦 Found Go $CURRENT_VERSION, upgrading to $REQUIRED_GO_VERSION..."
else
  echo "📦 Installing Go $REQUIRED_GO_VERSION..."
fi

# Download Go if not already present
if [ ! -f "/tmp/${GO_ARCHIVE}" ]; then
  echo "⬇️  Downloading Go ${REQUIRED_GO_VERSION}..."
  curl -L -o "/tmp/${GO_ARCHIVE}" "$GO_URL"
else
  echo "✅ Using cached Go archive"
fi

# Install Go
echo "📦 Installing Go to /usr/local/go..."
rm -rf /usr/local/go
tar -C /usr/local -xzf "/tmp/${GO_ARCHIVE}"

# Verify installation
if /usr/local/go/bin/go version | grep -q "$REQUIRED_GO_VERSION"; then
  echo "✅ Go $REQUIRED_GO_VERSION installed successfully"
else
  echo "❌ Go installation failed"
  exit 1
fi

# Clean up archive to save space (optional - comment out to keep for faster re-runs)
# rm -f "/tmp/${GO_ARCHIVE}"

echo "✅ Environment setup complete"

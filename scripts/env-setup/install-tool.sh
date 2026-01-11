#!/bin/bash
# Unified tool installer for splice development environment
# Usage: install-tool.sh <tmux|freeze>
#
# Detects the environment (Claude Code web, macOS, other) and installs
# the requested tool using the appropriate package manager.

set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 <tmux|freeze>"
  exit 1
fi

readonly TOOL="$1"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

case "$TOOL" in
  tmux)
    source "${SCRIPT_DIR}/install-tmux.sh"
    install_tmux
    ;;
  freeze)
    source "${SCRIPT_DIR}/install-freeze.sh"
    install_freeze
    ;;
  *)
    echo "❌ Unknown tool: $TOOL"
    echo "Supported tools: tmux, freeze"
    exit 1
    ;;
esac

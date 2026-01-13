# Claude Code Sandbox
# Pre-built image with Go toolchain and Claude Code CLI

FROM ubuntu:24.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install base dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    ca-certificates \
    gnupg \
    tmux \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 22.x (LTS) for Claude Code CLI
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.25.2
RUN curl -fsSL https://go.dev/dl/go1.25.2.linux-arm64.tar.gz | tar -C /usr/local -xzf -

# Create non-root user (no sudo access for security)
RUN useradd -m -s /bin/bash agent

# Set up environment for agent user
ENV PATH="/home/agent/.local/bin:/usr/local/go/bin:/home/agent/go/bin:${PATH}"
ENV GOPATH="/home/agent/go"
ENV HOME="/home/agent"
ENV TERM=xterm-256color

# Switch to non-root user for Claude Code installation
USER agent

# Install Claude Code CLI using official install script as agent user
# This installs the native binary with built-in auto-update support
RUN curl -fsSL https://claude.ai/install.sh | bash

# Pre-configure Claude Code to skip onboarding and permission dialogs
# ~/.claude.json = state file (onboarding, notifications, per-project settings)
# ~/.claude/settings.json = user settings (permissions mode, model)
#
# Key flags in .claude.json:
#   - hasCompletedOnboarding: skip initial setup wizard
#   - bypassPermissionsModeAccepted: skip "bypass permissions mode" notification
#   - projects[path].hasTrustDialogAccepted: skip workspace trust dialog (path-specific)
#   - projects[path].hasCompletedProjectOnboarding: skip project onboarding
RUN mkdir -p /home/agent/.claude && \
    echo '{"hasCompletedOnboarding":true,"theme":"dark","bypassPermissionsModeAccepted":true,"projects":{"/workspace":{"hasTrustDialogAccepted":true,"hasCompletedProjectOnboarding":true}}}' > /home/agent/.claude.json

# Set working directory
WORKDIR /workspace

# Default: keep container running
CMD ["sleep", "infinity"]

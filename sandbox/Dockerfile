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
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 22.x (LTS) for Claude Code CLI
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.25.2
RUN curl -fsSL https://go.dev/dl/go1.25.2.linux-arm64.tar.gz | tar -C /usr/local -xzf -

# Install Claude Code CLI globally
RUN npm install -g @anthropic-ai/claude-code

# Create non-root user
RUN useradd -m -s /bin/bash claude && \
    echo "claude ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Set up environment for claude user
ENV PATH="/usr/local/go/bin:/home/claude/go/bin:${PATH}"
ENV GOPATH="/home/claude/go"
ENV HOME="/home/claude"
ENV TERM=xterm-256color

# Pre-configure Claude Code to skip onboarding
RUN mkdir -p /home/claude/.claude && \
    echo '{"hasCompletedOnboarding":true,"hasSeenPermissionsNotice":true,"theme":"dark"}' > /home/claude/.claude/settings.json && \
    chown -R claude:claude /home/claude

# Switch to non-root user
USER claude

# Set working directory
WORKDIR /workspace

# Default: keep container running
CMD ["sleep", "infinity"]

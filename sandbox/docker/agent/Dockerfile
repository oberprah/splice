# Claude Code Sandbox Agent
# Go toolchain + Claude Code CLI in isolated environment

FROM ubuntu:24.04

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install base dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    git \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 22.x (required for Claude Code CLI)
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install Go (detect architecture)
ARG GO_VERSION=1.25.2
RUN ARCH=$(dpkg --print-architecture) && \
    case "$ARCH" in \
        amd64) GOARCH=amd64 ;; \
        arm64) GOARCH=arm64 ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" | tar -C /usr/local -xzf -

ENV PATH="/usr/local/go/bin:${PATH}"

# Install Codex CLI globally (as root)
RUN npm install -g @openai/codex

# Create non-root user
RUN useradd -m -u 1000 agent

# Install Claude Code as agent user
USER agent
RUN curl -fsSL https://claude.ai/install.sh | bash
ENV PATH="/home/agent/.local/bin:/home/agent/go/bin:${PATH}"
ENV GOPATH="/home/agent/go"
ENV HOME="/home/agent"
ENV TERM="xterm-256color"

# Pre-configure Claude Code (skip onboarding)
COPY --chown=agent:agent claude.json /home/agent/.claude.json

# Pre-configure Codex CLI to use internal proxy
COPY --chown=agent:agent codex-config.toml /home/agent/.codex/config.toml

WORKDIR /workspace

# Keep container running for interactive access
CMD ["sleep", "infinity"]

# Claude Code Sandbox Agent
# Go toolchain + Claude Code CLI + Codex CLI in isolated environment

FROM golang:1.25.2-bookworm

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install Node.js 22.x (required for Claude Code and Codex CLI)
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

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

# Mark /workspace as safe for git (volume has different ownership)
RUN git config --global --add safe.directory /workspace

# Pre-configure Claude Code (skip onboarding)
COPY --chown=agent:agent claude.json /home/agent/.claude.json

# Pre-configure Codex CLI to use internal proxy
COPY --chown=agent:agent codex-config.toml /home/agent/.codex/config.toml

WORKDIR /workspace

# Keep container running for interactive access
CMD ["sleep", "infinity"]

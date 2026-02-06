# Claude Code Sandbox Agent
# Go toolchain + Claude Code CLI + Codex CLI in isolated environment

FROM golang:1.25.2-bookworm

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install Node.js 24.x (required for Claude Code, Codex CLI, and OpenCode)
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install tmux (required for tape-runner tests)
RUN apt-get update \
    && apt-get install -y tmux \
    && rm -rf /var/lib/apt/lists/*

# Install Codex CLI globally (as root)
RUN npm install -g @openai/codex

# Symlink go to /usr/local/bin (Codex CLI uses a minimal PATH that includes this)
RUN ln -s /usr/local/go/bin/go /usr/local/bin/go

# Create non-root user
RUN useradd -m -u 1000 agent

# Switch to agent user and set up environment
USER agent
ENV HOME="/home/agent"
ENV TERM="xterm-256color"

# Set up Go environment
ENV GOPATH="/home/agent/go"
ENV PATH="/home/agent/go/bin:${PATH}"

# Install Claude Code CLI
RUN curl -fsSL https://claude.ai/install.sh | bash
ENV PATH="/home/agent/.local/bin:${PATH}"

# Install OpenCode
RUN curl -fsSL https://opencode.ai/install | bash
ENV PATH="/home/agent/.opencode/bin:${PATH}"

# Mark /workspace as safe for git (volume has different ownership)
RUN git config --global --add safe.directory /workspace

WORKDIR /workspace

# Keep container running for interactive access
CMD ["sleep", "infinity"]

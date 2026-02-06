# Claude Code Sandbox Agent
# Lightweight base + mise for language/tool management + AI agent CLIs

FROM debian:bookworm-slim

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Configure mise directories and install path
ENV MISE_DATA_DIR="/usr/local/mise"
ENV MISE_CACHE_DIR="/usr/local/mise/cache"
ENV MISE_INSTALL_PATH="/usr/local/bin/mise"
ENV PATH="/usr/local/mise/shims:$PATH"

# Install mise and essential build tools
RUN apt-get update \
    && apt-get install -y curl git ca-certificates build-essential tmux \
    && curl https://mise.run | sh \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user and mise directories with proper permissions
RUN useradd -m -u 1000 agent \
    && mkdir -p /usr/local/mise/cache \
    && chown -R agent:agent /usr/local/mise

# Switch to agent user and set up environment
USER agent
ENV HOME="/home/agent"
ENV TERM="xterm-256color"

# Install Node.js 24.x as agent user (required for AI agent CLIs)
RUN mise use -g node@24

# Install Codex CLI (as agent user)
RUN npm install -g @openai/codex

# Install Claude Code CLI
RUN curl -fsSL https://claude.ai/install.sh | bash
ENV PATH="/home/agent/.local/bin:${PATH}"

# Install OpenCode
RUN curl -fsSL https://opencode.ai/install | bash
ENV PATH="/home/agent/.opencode/bin:${PATH}"

# Install project-specific tools from .mise.toml
WORKDIR /workspace
COPY .mise.toml /workspace/.mise.toml
RUN mise trust && mise install -y

# Mark /workspace as safe for git (volume has different ownership)
RUN git config --global --add safe.directory /workspace

# Keep container running for interactive access
CMD ["sleep", "infinity"]

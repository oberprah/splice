#!/bin/bash
# Start the sandbox environment

set -e

cd "$(dirname "$0")"

# Build images
docker compose build

# Start both services (sandbox waits for litellm health check)
docker compose up -d

# Exec into sandbox and run claude
docker compose exec sandbox claude --dangerously-skip-permissions

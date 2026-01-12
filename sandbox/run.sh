#!/bin/bash
# Start the sandbox environment

set -e

cd "$(dirname "$0")"

# Build images
docker compose build

# Start litellm in background
docker compose up -d litellm

# Wait for litellm to be healthy
echo -n "Waiting for LiteLLM to be ready"
until docker compose exec -T litellm python -c "import urllib.request; urllib.request.urlopen('http://localhost:4000/health')" >/dev/null 2>&1; do
    echo -n "."
    sleep 2
done
echo " ready."

# Start sandbox in background
docker compose up -d sandbox

# Exec into sandbox and run claude
docker compose exec sandbox claude --dangerously-skip-permissions

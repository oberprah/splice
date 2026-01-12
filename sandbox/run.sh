#!/bin/bash
# Start the sandbox environment

set -e

cd "$(dirname "$0")"

# Build and start
docker compose up -d --build

# Attach to sandbox
docker compose exec sandbox bash

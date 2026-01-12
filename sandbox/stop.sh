#!/bin/bash
# Stop the sandbox environment

cd "$(dirname "$0")"
docker compose down

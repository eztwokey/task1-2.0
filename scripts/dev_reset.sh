#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
echo "Stopping and removing compose stack..."
docker compose -f deploy/docker-compose.yml down -v --remove-orphans || true
echo "Pruning dangling images and volumes (careful)..."
docker system prune -f || true
docker volume prune -f || true
echo "Done."

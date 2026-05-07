#!/usr/bin/env bash
set -euo pipefail

# Aplica migraciones SQL directo via psql en container postgres.
COMPOSE="docker compose -f deploy/docker-compose.yml"
DB_USER="${POSTGRES_USER:-tama}"
DB_NAME="${POSTGRES_DB:-tama}"

for svc in auth pet social; do
  for f in services/$svc/migrations/*.up.sql; do
    echo "==> $svc :: $(basename $f)"
    $COMPOSE exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" < "$f"
  done
done

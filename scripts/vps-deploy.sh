#!/usr/bin/env bash
# Deploy / update producción en VPS. Idempotente.
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/tamagotchi}"
cd "$APP_DIR"

echo "==> Pull latest"
git fetch origin
git reset --hard origin/main

echo "==> Build images"
cd deploy/prod
docker compose --env-file ../../.env build --pull

echo "==> Up stack"
docker compose --env-file ../../.env up -d

echo "==> Waiting Postgres healthy"
for i in {1..30}; do
  if docker compose exec -T postgres pg_isready -U "$(grep POSTGRES_USER ../../.env | cut -d= -f2)" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "==> Apply migrations"
DB_USER=$(grep POSTGRES_USER ../../.env | cut -d= -f2)
DB_NAME=$(grep POSTGRES_DB ../../.env | cut -d= -f2)
for svc in auth pet social; do
  for f in /migrations/$svc/*.up.sql; do
    docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -f "$f" || true
  done 2>/dev/null
  for f in ../../services/$svc/migrations/*.up.sql; do
    docker compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" < "$f" || true
  done
done

echo "==> Status"
docker compose ps

echo
echo "Listo. Verifica:"
echo "  docker compose logs -f"
echo "  curl https://\$API_DOMAIN/healthz"

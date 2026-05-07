#!/usr/bin/env bash
set -euo pipefail

cp -n .env.example .env || true

echo "==> Tidy Go modules"
for svc in auth pet social notif gateway; do
  (cd services/$svc && go mod tidy)
done

echo "==> Install web deps"
(cd web && npm install)

echo "==> Compose up"
docker compose -f deploy/docker-compose.yml up -d --build

echo "==> Run migrations (once postgres ready)"
sleep 5
for svc in auth pet social; do
  docker compose -f deploy/docker-compose.yml exec -T postgres \
    psql -U "${POSTGRES_USER:-tama}" -d "${POSTGRES_DB:-tama}" \
    -f /app/migrations/001_init.up.sql || true
done

echo "Done. Frontend: cd web && npm run dev"

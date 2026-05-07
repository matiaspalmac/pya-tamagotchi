#!/usr/bin/env bash
# Despliegue completo Fly.io - corre desde raíz repo: bash scripts/fly-deploy.sh
set -euo pipefail

ORG="${FLY_ORG:-personal}"
REGION="${FLY_REGION:-scl}"
DB_APP="tama-db"
REDIS_NAME="tama-redis"

require() { command -v "$1" >/dev/null || { echo "falta $1"; exit 1; }; }
require flyctl

echo "==> 1. Login (skip si ya logueado)"
flyctl auth whoami || flyctl auth login

echo "==> 2. Postgres cluster"
flyctl postgres list | grep -q "$DB_APP" || \
  flyctl postgres create --name "$DB_APP" --region "$REGION" \
    --initial-cluster-size 1 --vm-size shared-cpu-1x --volume-size 1

echo "==> 3. Crear apps (idempotente)"
for app in tama-auth tama-pet tama-social tama-notif tama-gateway tama-web; do
  flyctl apps list | grep -q "$app" || flyctl apps create "$app" --org "$ORG"
done

echo "==> 4. Attach Postgres a servicios que lo usan"
for app in tama-auth tama-pet tama-social; do
  flyctl postgres attach "$DB_APP" -a "$app" --yes || echo "(ya attached $app)"
done

echo "==> 5. Upstash Redis (compartido pet+notif+social)"
flyctl ext list -a tama-pet | grep -q upstash-redis || \
  flyctl ext create upstash-redis -a tama-pet --name "$REDIS_NAME" --region "$REGION"

REDIS_URL=$(flyctl ssh console -a tama-pet -C "printenv REDIS_URL" 2>/dev/null || echo "")
if [ -z "$REDIS_URL" ]; then
  echo "OBTÉN REDIS_URL con: flyctl ext show $REDIS_NAME -a tama-pet"
  echo "Y exporta REDIS_URL=... antes de re-correr"
  exit 1
fi

# Parse host/port de redis://default:pass@host:port
REDIS_HOST=$(echo "$REDIS_URL" | sed -E 's#.*@([^:]+):.*#\1#')
REDIS_PORT=$(echo "$REDIS_URL" | sed -E 's#.*:([0-9]+).*#\1#')
REDIS_PASS=$(echo "$REDIS_URL" | sed -E 's#redis://default:([^@]+)@.*#\1#')

echo "==> 6. JWT secret común"
JWT_SECRET="${JWT_SECRET:-$(openssl rand -hex 32)}"

echo "==> 7. Set secrets en cada servicio"
for app in tama-auth tama-pet tama-social tama-notif tama-gateway; do
  flyctl secrets set -a "$app" \
    JWT_SECRET="$JWT_SECRET" \
    REDIS_HOST="$REDIS_HOST" \
    REDIS_PORT="$REDIS_PORT" \
    REDIS_PASSWORD="$REDIS_PASS" \
    --stage
done

echo "==> 8. Deploy en orden (data svcs primero)"
flyctl deploy -c deploy/fly/fly.auth.toml    -a tama-auth    --remote-only
flyctl deploy -c deploy/fly/fly.pet.toml     -a tama-pet     --remote-only
flyctl deploy -c deploy/fly/fly.social.toml  -a tama-social  --remote-only
flyctl deploy -c deploy/fly/fly.notif.toml   -a tama-notif   --remote-only
flyctl deploy -c deploy/fly/fly.gateway.toml -a tama-gateway --remote-only

GATEWAY_URL="https://tama-gateway.fly.dev"

echo "==> 9. Deploy frontend con VITE_API_URL"
flyctl deploy -c deploy/fly/fly.web.toml -a tama-web --remote-only \
  --build-arg VITE_API_URL="$GATEWAY_URL"

echo "==> 10. Migraciones Postgres"
for svc in auth pet social; do
  echo "-- migrating $svc --"
  flyctl postgres connect -a "$DB_APP" < "services/$svc/migrations/001_init.up.sql" || true
done

echo
echo "Done."
echo "Frontend:  https://tama-web.fly.dev"
echo "API:       $GATEWAY_URL"

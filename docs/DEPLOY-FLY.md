# Deploy Fly.io

## Topología

```
                 Internet
                    │
                    ▼
         ┌──────────────────────┐
         │  tama-web (nginx)    │  Vite build estático
         │  https://tama-web…   │
         └──────────┬───────────┘
                    │ fetch / WS
                    ▼
         ┌──────────────────────┐
         │  tama-gateway        │  público :443
         │  https://tama-gw…    │  JWT + rate limit + proxy
         └──┬─────┬─────┬─────┬─┘
            │     │     │     │   red privada .flycast
   ┌────────▼┐ ┌──▼──┐ ┌▼────┐ ┌▼─────┐
   │tama-auth│ │pet  │ │soc  │ │notif │
   └────┬────┘ └─┬───┘ └─┬───┘ └──┬───┘
        │        │       │        │
   ┌────▼────────▼───────▼────┐ ┌─▼────────┐
   │   Fly Postgres tama-db   │ │ Upstash  │
   └──────────────────────────┘ │ Redis    │
                                └──────────┘
```

## Prerequisitos

- Cuenta Fly.io + tarjeta (free tier requiere tarjeta verificación)
- `flyctl` instalado: https://fly.io/docs/flyctl/install/
- `openssl` (genera JWT secret)

## Deploy automático

```bash
bash scripts/fly-deploy.sh
```

Hace todo: crea apps, Postgres, Redis (Upstash ext), set secrets, deploy en orden, migraciones.

## Deploy manual paso a paso

### 1. Login + Postgres

```bash
flyctl auth login
flyctl postgres create --name tama-db --region scl \
  --initial-cluster-size 1 --vm-size shared-cpu-1x --volume-size 1
```

### 2. Crear apps

```bash
for app in tama-auth tama-pet tama-social tama-notif tama-gateway tama-web; do
  flyctl apps create $app
done
```

### 3. Attach Postgres a servicios con DB

```bash
flyctl postgres attach tama-db -a tama-auth
flyctl postgres attach tama-db -a tama-pet
flyctl postgres attach tama-db -a tama-social
```

Esto inyecta `DATABASE_URL` automáticamente. Mapea a env vars Postgres en código (ya hecho via secrets).

### 4. Redis Upstash

```bash
flyctl ext create upstash-redis -a tama-pet --name tama-redis --region scl
flyctl ext show tama-redis -a tama-pet  # copia REDIS_URL
```

Comparte la misma instancia Redis con `tama-notif` y `tama-social`:

```bash
flyctl secrets set REDIS_HOST=... REDIS_PORT=... REDIS_PASSWORD=... -a tama-notif
flyctl secrets set REDIS_HOST=... REDIS_PORT=... REDIS_PASSWORD=... -a tama-social
flyctl secrets set REDIS_HOST=... REDIS_PORT=... REDIS_PASSWORD=... -a tama-pet
```

### 5. JWT secret común

```bash
JWT=$(openssl rand -hex 32)
for app in tama-auth tama-pet tama-social tama-notif tama-gateway; do
  flyctl secrets set JWT_SECRET=$JWT -a $app
done
```

### 6. Deploy servicios

```bash
flyctl deploy -c deploy/fly/fly.auth.toml    -a tama-auth
flyctl deploy -c deploy/fly/fly.pet.toml     -a tama-pet
flyctl deploy -c deploy/fly/fly.social.toml  -a tama-social
flyctl deploy -c deploy/fly/fly.notif.toml   -a tama-notif
flyctl deploy -c deploy/fly/fly.gateway.toml -a tama-gateway
```

### 7. Deploy frontend

```bash
flyctl deploy -c deploy/fly/fly.web.toml -a tama-web \
  --build-arg VITE_API_URL=https://tama-gateway.fly.dev
```

### 8. Migraciones

```bash
flyctl postgres connect -a tama-db < services/auth/migrations/001_init.up.sql
flyctl postgres connect -a tama-db < services/pet/migrations/001_init.up.sql
flyctl postgres connect -a tama-db < services/social/migrations/001_init.up.sql
```

## Red privada `.flycast`

Servicios internos comunican via `tama-auth.flycast:8081`, etc. Solo `gateway` y `web` exponen público (`force_https`). Esto está configurado en cada `fly.toml` con `[[services]]` (no `[http_service]`) para los internos.

## Free tier costos

| Recurso | Free | Costo si excede |
|---------|------|-----------------|
| 6 apps × 256MB shared | 3 incluidas | ~$1.94/mes c/u extra |
| Postgres dev 1GB | 1 free | $1.94/mes 256MB |
| Upstash Redis 10k cmd/día | free | $0.20 / 100k cmd |
| Egress 160GB/mes | free | $0.02/GB |

**5 servicios + frontend = excede 3 free.** Costo estimado: **~$6-10/mes**.

Ahorro: consolidar a 3 apps (gateway+notif, auth+social, pet) = $0/mes.

## Logs y debug

```bash
flyctl logs -a tama-pet
flyctl status -a tama-gateway
flyctl ssh console -a tama-auth
flyctl machine list -a tama-pet
flyctl scale count 2 -a tama-gateway   # escalar
```

## Updates

```bash
flyctl deploy -c deploy/fly/fly.pet.toml -a tama-pet
```

CI/CD futuro: GitHub Actions con `superfly/flyctl-actions@v1`.

## Rollback

```bash
flyctl releases -a tama-pet
flyctl releases rollback <version> -a tama-pet
```

## Custom domain

```bash
flyctl certs create tudominio.cl -a tama-web
flyctl certs create api.tudominio.cl -a tama-gateway
# añadir CNAME en DNS apuntando a *.fly.dev
```

Re-deploy frontend con `VITE_API_URL=https://api.tudominio.cl`.

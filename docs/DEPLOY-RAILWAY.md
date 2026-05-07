# Deploy Railway

## Topología

```
Railway project: pya-tamagotchi
├── Postgres (plugin)
├── Redis    (plugin)
├── Service: auth      → Dockerfile services/auth
├── Service: pet       → Dockerfile services/pet
├── Service: social    → Dockerfile services/social
├── Service: notif     → Dockerfile services/notif
├── Service: gateway   → Dockerfile services/gateway  [public]
└── Service: web       → Dockerfile web               [public]
```

## Costo

- Hobby plan: **$5/mes flat** + uso por encima
- $5 crédito incluido cubre: ~512MB RAM × varios servicios pequeños 24/7
- Postgres + Redis plugins cuentan al uso pero baratos en hobby
- Estimado este proyecto: **$5-10/mes**

## Networking interno

Servicios mismo proyecto comunican vía:
```
http://${{auth.RAILWAY_PRIVATE_DOMAIN}}:8081
http://${{pet.RAILWAY_PRIVATE_DOMAIN}}:8082
```

Solo `gateway` y `web` exponen público (Railway genera dominio `*.up.railway.app`).

## Setup vía Dashboard (recomendado primera vez)

### 1. Crear proyecto

https://railway.com/new → "Deploy from GitHub repo" → autoriza → selecciona `pya-tamagotchi`.

Railway detecta `railway.json` en cada subpath. Crea **un servicio por defecto** apuntando a la raíz. Lo eliminamos y creamos 6 manuales:

### 2. Add Postgres + Redis plugins

Project → "+ New" → Database → PostgreSQL.
Project → "+ New" → Database → Redis.

Esto crea variables `DATABASE_URL`, `REDIS_URL`, `PGHOST`, `PGPORT`, etc, accesibles vía `${{Postgres.PGHOST}}`.

### 3. Crear cada servicio

Por cada uno (auth, pet, social, notif, gateway, web):

1. Project → "+ New" → "GitHub Repo" → mismo repo
2. Settings → Root Directory: `services/auth` (o `web`)
3. Settings → Watch Paths: `services/auth/**` (auto-redeploy solo si cambia)
4. Variables: agregar (ver abajo)
5. Networking → "Generate Domain" SOLO para `gateway` y `web`

### 4. Variables por servicio

**Common (todos los Go svcs):**
```
JWT_SECRET=<generado openssl rand -hex 32, mismo en los 5>
LOG_LEVEL=info
POSTGRES_HOST=${{Postgres.PGHOST}}
POSTGRES_PORT=${{Postgres.PGPORT}}
POSTGRES_USER=${{Postgres.PGUSER}}
POSTGRES_PASSWORD=${{Postgres.PGPASSWORD}}
POSTGRES_DB=${{Postgres.PGDATABASE}}
REDIS_HOST=${{Redis.REDISHOST}}
REDIS_PORT=${{Redis.REDISPORT}}
REDIS_PASSWORD=${{Redis.REDISPASSWORD}}
```

**auth (8081):**
```
SERVICE_PORT=8081
JWT_EXPIRE_MIN=60
JWT_REFRESH_DAYS=30
PORT=8081
```

**pet (8082):**
```
SERVICE_PORT=8082
TICK_INTERVAL_SEC=30
PORT=8082
```

**social (8083):**
```
SERVICE_PORT=8083
PORT=8083
```

**notif (8084):**
```
SERVICE_PORT=8084
PORT=8084
```

**gateway (8080) — público:**
```
SERVICE_PORT=8080
PORT=8080
AUTH_URL=http://${{auth.RAILWAY_PRIVATE_DOMAIN}}:8081
PET_URL=http://${{pet.RAILWAY_PRIVATE_DOMAIN}}:8082
SOCIAL_URL=http://${{social.RAILWAY_PRIVATE_DOMAIN}}:8083
NOTIF_URL=http://${{notif.RAILWAY_PRIVATE_DOMAIN}}:8084
CORS_ORIGINS=https://<tu-web>.up.railway.app
```

**web (80) — público:**
Build arg en Dockerfile:
```
VITE_API_URL=https://<tu-gateway>.up.railway.app
```

Configurar en Settings → Build → Build Args.

### 5. Migraciones Postgres

Una vez Postgres up:
```bash
railway link    # selecciona proyecto
railway connect Postgres
# psql shell:
\i services/auth/migrations/001_init.up.sql
\i services/pet/migrations/001_init.up.sql
\i services/social/migrations/001_init.up.sql
```

O via Railway CLI:
```bash
railway run --service auth psql $DATABASE_URL -f services/auth/migrations/001_init.up.sql
```

## Setup vía CLI (faster, automatable)

```bash
railway login
railway init pya-tamagotchi
railway add --plugin postgresql
railway add --plugin redis

# Por servicio
cd services/auth
railway add --service auth
railway up --service auth
# repite pa cada svc...
```

## Re-deploys

Push a `main` → Railway auto-deploya servicios cuyos `Watch Paths` matchearon cambios.

## Logs

Dashboard → Service → "Deployments" → click → Logs.

CLI:
```bash
railway logs --service pet
```

## Rollback

Dashboard → Service → Deployments → "Redeploy" en versión anterior.

## Custom domain

Service → Settings → Networking → "Custom Domain" → ingresa dominio + agregar CNAME en DNS.

## Diferencias vs Fly.io

| | Fly | Railway |
|--|-----|---------|
| Free tier | $0 (eliminado) | $5 crédito Hobby |
| Setup | CLI heavy | Dashboard friendly |
| Networking interno | `.flycast` | `RAILWAY_PRIVATE_DOMAIN` |
| WS | nativo | nativo |
| Multi-service repo | apps separadas | servicios mismo project |
| Postgres | unmanaged DIY o $$ managed | plugin 1-click |
| Redis | extension Upstash | plugin 1-click |

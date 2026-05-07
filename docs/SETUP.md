# Setup local

## Requisitos

- Docker Desktop
- Go 1.23+
- Node 20+
- Make (opcional, ayuda)

## Pasos

```bash
git clone <repo-url>
cd tamagotchi
cp .env.example .env

# Levanta Postgres + Redis + servicios
make up

# Aplica migraciones (primera vez)
bash scripts/migrate.sh

# Frontend dev
make web-install
make web   # http://localhost:5173
```

## Verificación

```bash
curl http://localhost:8080/healthz
# ok

curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"a@b.cl","username":"safe","password":"pass1234"}'
```

## Logs

```bash
make logs                                  # todos
docker compose -f deploy/docker-compose.yml logs -f pet
```

## Reset DB

```bash
docker compose -f deploy/docker-compose.yml down -v
make up
bash scripts/migrate.sh
```

## Trabajar en un servicio individual

```bash
cd services/pet
go mod tidy
go test ./...
go run ./cmd
```

Para que corra sin Docker, exporta env vars del `.env` y asegúrate Postgres/Redis estén accesibles en `localhost`.

## Frontend solo

```bash
cd web
npm install
npm run dev
```

Vite proxy apunta a `localhost:8080` (gateway).

## Troubleshooting

- **`relation "users" does not exist`** → corre `bash scripts/migrate.sh`
- **`connection refused redis`** → `docker compose ps`, verifica Redis healthy
- **Frontend 401 después de un rato** → access token expiró (60min), implementar refresh en cliente (Sprint 2 backlog)
- **CORS errors** → revisar gateway CORS config y `vite.config.ts` proxy

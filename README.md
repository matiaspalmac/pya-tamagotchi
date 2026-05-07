# Tamagotchi Multiplayer Web

Proyecto comunitario de aprendizaje. Microservicios Go + React + Postgres + Redis + WebSockets.

## Stack

- **Backend:** Go 1.23, chi, sqlx, gorilla/websocket
- **Frontend:** React 18 + Vite + TypeScript + Zustand + Tailwind
- **Datos:** PostgreSQL 16, Redis 7
- **Infra:** Docker Compose (dev), Fly.io (deploy), GitHub Actions (CI)

## Servicios

| Servicio | Puerto | Responsabilidad |
|----------|--------|-----------------|
| gateway  | 8080   | Reverse proxy, auth middleware, rate limit |
| auth     | 8081   | Registro, login, JWT |
| pet      | 8082   | Mascotas, stats, tick loop, acciones |
| social   | 8083   | Amigos, visitas, regalos |
| notif    | 8084   | WebSocket hub, eventos realtime |

## Quick start

```bash
make up          # levanta Postgres + Redis + servicios
make migrate     # corre migraciones
make seed        # data demo opcional
make web         # frontend en http://localhost:5173
```

API gateway: http://localhost:8080
WebSocket:   ws://localhost:8080/ws

## Estructura

```
tamagotchi/
├── services/        # microservicios Go
│   ├── auth/
│   ├── pet/
│   ├── social/
│   ├── notif/
│   └── gateway/
├── pkg/             # libs compartidas (jwt, http, log, db clients)
├── web/             # React app
├── deploy/          # Fly.io, nginx
├── docs/            # spec, ADRs, diagramas
└── .github/workflows/
```

## Roles equipo

| Persona | Rol | Servicios owned |
|---------|-----|-----------------|
| Safe    | Backend Lead | pet, notif |
| Dei     | Arquitecto + Backend | gateway, auth, infra |
| P3      | Frontend Lead | web/ |
| P4      | Backend Jr | social |
| P5      | DevOps/QA | CI, deploy, tests |

3 personas: Safe = pet+notif, Dei = gateway+auth+infra, P3 = web.

## Reglas

- Branches: `feat/<svc>-<desc>`, `fix/<svc>-<desc>`
- Conventional commits (`feat(pet): add tick loop`)
- PR review obligatorio (1 approver mín)
- Tests requeridos en lógica game loop y auth
- IA: dudas/docs sí. Copy-paste features completos no. Demo explica tu PR.
- Daily async Discord: ayer / hoy / blockers
- Demo viernes 30min cierre sprint

## Roadmap

- **Sprint 1 (sem 1-2):** Foundation — auth + skeleton + Docker
- **Sprint 2 (sem 3-4):** Core pet — CRUD + tick + UI básica
- **Sprint 3 (sem 5-6):** Realtime + social — WS + amigos
- **Sprint 4 (sem 7-8):** Polish + deploy — regalos + stages + Fly.io

Ver `docs/SPEC.md` para detalle.

## Deploy

Plataforma: **DigitalOcean Droplet** (vía GitHub Student Pack $200 crédito).

Stack: VPS Ubuntu + Docker Compose + Caddy (HTTPS auto) + dominio Namecheap .me free.

```bash
# En VPS Ubuntu 24.04 fresco:
curl -fsSL https://raw.githubusercontent.com/matiaspalmac/pya-tamagotchi/main/scripts/vps-bootstrap.sh | bash
nano /opt/tamagotchi/.env   # secrets + dominios
bash /opt/tamagotchi/scripts/vps-deploy.sh
```

Detalles paso a paso: `docs/DEPLOY-VPS.md`.

## Licencia

MIT

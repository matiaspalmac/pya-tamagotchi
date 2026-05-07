# Roadmap

## Sprint 1 — Foundation (sem 1-2)

**Meta:** infra base + auth funcionando.

- [ ] Repo creado, branch protection main, CODEOWNERS
- [ ] Docker Compose levanta Postgres + Redis
- [ ] auth-svc: register/login/refresh/me
- [ ] Migraciones auth corren
- [ ] gateway enruta /auth/*
- [ ] Frontend: páginas Login + Register conectadas
- [ ] CI verde (build + test + web build)
- [ ] Demo: usuario crea cuenta, hace login, ve dashboard vacío

**Owners:**
- Dei: gateway, auth-svc, infra
- Safe: revisión arquitectura
- P3: web Login/Register/Dashboard skeleton

## Sprint 2 — Core pet (sem 3-4)

**Meta:** crear mascota + tick + acciones básicas.

- [ ] pet-svc: CRUD pet
- [ ] Tick lógica + tests (cobertura >70% en game/)
- [ ] Acciones feed/play/sleep/heal con cooldowns Redis
- [ ] Eventos persistidos en pet_events
- [ ] Frontend PetCard con stats animadas
- [ ] Dashboard lista mascotas
- [ ] Demo: crear mascota, ver decay 5min, alimentar, ver stat subir

**Owners:**
- Safe: pet-svc completo
- P3: PetCard + Dashboard
- Dei: review

## Sprint 3 — Realtime + social (sem 5-6)

**Meta:** WebSockets + amigos.

- [ ] notif-svc: WS hub + Redis pub/sub
- [ ] gateway WS proxy
- [ ] Frontend conecta WS, refresca en tick
- [ ] social-svc: friend request/accept/list
- [ ] social-svc: gifts send/inbox/claim
- [ ] Frontend: tab Amigos, enviar regalo
- [ ] Demo: dos usuarios en paralelo, uno envía regalo, otro lo ve aparecer en vivo

**Owners:**
- Safe: notif-svc + WS frontend integration
- P4 / Dei: social-svc
- P3: UI amigos + regalos

## Sprint 4 — Polish + deploy (sem 7-8)

**Meta:** evolución + deploy público.

- [ ] Lógica evolución stages (egg→baby→teen→adult→elder)
- [ ] Sprites por stage
- [ ] Tests integración (pet + social)
- [ ] Tests E2E Playwright (registro→pet→feed)
- [ ] Deploy Fly.io (gateway + servicios)
- [ ] Postgres + Redis managed (Neon/Upstash)
- [ ] Domain custom + HTTPS
- [ ] Demo pública en comunidad

**Owners:**
- P5 / Dei: deploy
- Safe: evolución + tests
- P3: sprites + UX final

## Backlog

- Marketplace gifts (compra coins)
- Leaderboard mascotas más viejas
- Logros / badges
- Mobile responsive completo
- i18n (en/es)
- Modo "guardería" (cuidar mascota amigo offline)
- Mascota multi-stage cosmetics
- Notificaciones web push
- Métricas Prometheus + Grafana

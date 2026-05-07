# Architecture

## Diagrama

```
                    ┌─────────────┐
                    │   Browser   │
                    │  React+Vite │
                    └──────┬──────┘
                           │ HTTPS / WSS
                    ┌──────▼──────┐
                    │   Gateway   │  :8080
                    │ chi + proxy │  CORS, JWT verify, rate limit
                    └──┬──┬──┬──┬─┘
        ┌──────────────┘  │  │  └──────────────┐
        │                 │  │                 │
   ┌────▼────┐      ┌─────▼──▼────┐      ┌─────▼────┐      ┌──────────┐
   │  auth   │      │     pet     │      │  social  │      │  notif   │
   │  :8081  │      │    :8082    │      │  :8083   │      │  :8084   │
   └────┬────┘      └─────┬───┬───┘      └────┬─────┘      └────┬─────┘
        │                 │   │               │                 │
        │           ┌─────┘   └─────┐         │                 │
        ▼           ▼               ▼         ▼                 ▼
   ┌────────────────────────────────────────────┐         ┌──────────┐
   │              PostgreSQL 16                 │         │ Redis 7  │
   │  users, pets, friendships, gifts, events   │◄────────┤ pub/sub  │
   └────────────────────────────────────────────┘         │ stats    │
                                                          └──────────┘
```

## Decisiones (ADR resumido)

### ADR-001: Microservicios vs monolito
**Decisión:** microservicios desde día 1.
**Razón:** objetivo aprendizaje. Dei quiere microservicios. Trade-off: más infra inicial.

### ADR-002: Postgres único compartido
**Decisión:** un Postgres, schemas separados por servicio (`auth.users`, `pet.pets`...).
**Razón:** simplicidad MVP. Migrar a DB-por-servicio si crece.

### ADR-003: Redis para estado caliente pet
**Decisión:** Redis HSET por mascota, flush periódico Postgres.
**Razón:** tick loop alta frecuencia, evita hotspot Postgres.

### ADR-004: WS centralizado en notif-svc
**Decisión:** solo notif maneja WS. Otros servicios publican a Redis pub/sub.
**Razón:** desacopla, escala notif independiente.

### ADR-005: JWT stateless
**Decisión:** HS256, sin sesión server-side.
**Razón:** simple, escala. Refresh tokens en DB.

## Flujo: feed pet

```
Browser → POST /pets/:id/feed (gateway)
       → JWT verify (gateway middleware)
       → proxy → pet-svc
       → pet-svc: lock Redis pet:<id>, validate cooldown
       → update Redis stats
       → INSERT events
       → publish Redis "pet:event" {type:"fed", pet_id}
       → response 200 {pet}
notif-svc ← subscribe "pet:event"
         → push WS a suscriptores
```

## Escalabilidad futura

- pet-svc tick loop: shard por pet_id hash, líder por shard (Redis lock)
- notif-svc: horizontal, sticky sessions o broker compartido (NATS)
- Postgres: read replicas
- CDN frontend

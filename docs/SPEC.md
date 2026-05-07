# Tamagotchi Multiplayer — Spec

## Visión

Mascota virtual persistente con stats que decaen en tiempo real. Usuarios cuidan, visitan amigos, envían regalos. Mascotas evolucionan según cuidado.

## Modelo dominio

### User
- id (uuid), email, username, password_hash, created_at

### Pet
- id, owner_id, name, species
- hunger (0-100), happy (0-100), energy (0-100), health (0-100)
- stage: egg | baby | teen | adult | elder
- born_at, last_tick_at, died_at (nullable)
- xp (int), evolution_threshold por stage

### Friendship
- user_a_id, user_b_id, status (pending|accepted), created_at

### Gift
- id, from_user_id, to_user_id, type (food|toy|medicine), claimed (bool), created_at

### Event (audit/feed)
- id, pet_id, type, payload jsonb, created_at

## Reglas juego

### Tick
- Intervalo: 30s configurable
- Por tick: hunger -= 2, happy -= 1, energy -= 1
- Si hunger < 20 → health -= 1
- Si health == 0 → died_at = now, stage frozen

### Acciones
- **feed:** hunger += 30 (cap 100), happy += 5
  - cooldown 60s, requiere food gift o gratis 1/h
- **play:** happy += 25, energy -= 15
  - cooldown 30s, requiere energy >= 20
- **sleep:** energy += 50 sobre 5min, no acciones durante sleep
- **heal:** health += 50, requiere medicine gift

### Evolución
- baby → teen: 24h vida + xp >= 100
- teen → adult: 72h vida + xp >= 500
- adult → elder: 7d vida
- xp gana por acciones bien-timing y health alta

## API

### auth-svc (8081)
```
POST   /auth/register   {email, username, password} → {user, tokens}
POST   /auth/login      {email, password}           → {user, tokens}
POST   /auth/refresh    {refresh_token}             → {tokens}
GET    /auth/me                                     → {user}    [JWT]
```

### pet-svc (8082)
```
POST   /pets              {name, species}     → Pet      [JWT]
GET    /pets/mine                              → [Pet]    [JWT]
GET    /pets/:id                               → Pet      [JWT]
POST   /pets/:id/feed                          → Pet      [JWT]
POST   /pets/:id/play                          → Pet      [JWT]
POST   /pets/:id/sleep                         → Pet      [JWT]
POST   /pets/:id/heal                          → Pet      [JWT]
GET    /pets/:id/events?limit=50               → [Event]  [JWT]
```

### social-svc (8083)
```
POST   /friends/request   {username}     → Friendship   [JWT]
POST   /friends/accept    {request_id}   → Friendship   [JWT]
GET    /friends                          → [User+Pet]   [JWT]
POST   /gifts             {to, type}     → Gift         [JWT]
GET    /gifts/inbox                      → [Gift]       [JWT]
POST   /gifts/:id/claim                  → Gift         [JWT]
GET    /feed                             → [Event]      [JWT]
```

### notif-svc (8084)
```
WS     /ws            (auth via ?token=JWT)
```
Mensajes server→client:
```json
{"type":"pet.tick",       "pet_id":"...", "stats":{...}}
{"type":"pet.evolved",    "pet_id":"...", "stage":"teen"}
{"type":"pet.died",       "pet_id":"..."}
{"type":"gift.received",  "from":"...",   "gift":{...}}
{"type":"friend.request", "from":"..."}
```
Cliente→server:
```json
{"type":"subscribe", "pet_ids":["..."]}
{"type":"ping"}
```

### gateway (8080)
- Enruta `/auth/*` → auth, `/pets/*` → pet, `/friends/*` `/gifts/*` `/feed` → social, `/ws` → notif
- Middleware: CORS, request-id, JWT verify (excepto /auth/login /auth/register)
- Rate limit: 100 req/min por IP

## Realtime flow

1. Cliente login → recibe JWT
2. Conecta WS `/ws?token=...` → notif-svc valida JWT
3. Cliente subscribe pet_ids
4. pet-svc tick loop publica a Redis `pet:tick` channel
5. notif-svc lee Redis pub/sub → push a clientes suscritos
6. Acciones (feed/play) publican `pet:event`

## Persistencia stats

- Estado caliente en Redis: `pet:<id>:stats` (HSET hunger/happy/etc)
- Tick loop actualiza Redis cada 30s
- Flush a Postgres cada 5min (snapshot) y on-action
- En cold start: lee Postgres → rebuild Redis con catch-up por (now - last_tick_at)

## Seguridad

- Password: bcrypt cost 12
- JWT: HS256, secret 32b mín, expire 60min, refresh 30d
- Rate limit gateway
- CORS whitelist explícita
- Validación input en cada handler (go-playground/validator)
- SQL: solo prepared statements / sqlx
- WS: token en query (HTTPS only en prod), ping/pong 30s

## Tests

- Unit: lógica tick, evolución, cooldowns
- Integration: auth flow, pet CRUD, WS subscribe→event
- E2E (Sprint 4): Playwright registro→crear pet→feed→ver tick

## Métricas (futuro)

- Prometheus: req_count, req_duration, ws_connections, ticks_per_sec
- Grafana dashboard

## Open questions

- ¿Mascota muere permanente o resucita? → MVP: muere permanente, user crea nueva
- ¿Multi-mascota por user? → MVP: 1 activa, archivo después
- ¿Marketplace gifts? → Sprint 5+

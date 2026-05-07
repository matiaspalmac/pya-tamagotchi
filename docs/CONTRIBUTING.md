# Contributing

## Workflow

1. Asigna issue antes de empezar
2. Branch: `feat/<svc>-<desc>` desde `main`
3. Commits: Conventional Commits
   - `feat(pet): add tick loop`
   - `fix(auth): jwt expiry off-by-one`
   - `docs(spec): clarify evolution rules`
   - `test(pet): cover feed cooldown`
4. Push + PR a `main`
5. 1 approver mín, CI verde
6. Squash merge

## Estilo Go

- `gofmt` obligatorio
- `golangci-lint run` antes PR
- Errors wrapped: `fmt.Errorf("feed pet: %w", err)`
- No panic en handlers
- Context propagado siempre
- Tests `_test.go` mismo package

## Estilo TS/React

- ESLint + Prettier configurados
- Componentes funcionales + hooks
- Zustand store por dominio
- Tipado estricto (no `any`)

## Tests mínimos

- auth-svc: register, login, refresh, JWT verify
- pet-svc: tick math, cooldowns, evolución thresholds, feed/play/sleep/heal
- social-svc: friend request flow, gift claim
- notif-svc: WS auth, subscribe filter
- gateway: routing, JWT middleware

## Code review checklist

- [ ] Tests añadidos/actualizados
- [ ] Sin secrets en código
- [ ] Errors propagados con contexto
- [ ] Validación input en handler
- [ ] Migraciones reversibles (up + down)
- [ ] README/SPEC actualizado si cambia API
- [ ] No copy-paste IA sin entender — explica oral en demo

## Política IA

- ✅ Preguntar conceptos, dudas docs, debugging local
- ✅ Generar boilerplate (structs, mocks)
- ❌ Pegar features completas sin entender
- ❌ Tests generados sin verificar que prueban algo real

Demo viernes: tu PR, tú lo explicas. Si no puedes, no era tuyo.

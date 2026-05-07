.PHONY: up down logs build migrate seed test lint web fmt tidy

COMPOSE=docker compose -f deploy/docker-compose.yml

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

build:
	$(COMPOSE) build

migrate:
	$(COMPOSE) exec auth   /app/migrate up
	$(COMPOSE) exec pet    /app/migrate up
	$(COMPOSE) exec social /app/migrate up

seed:
	$(COMPOSE) exec pet /app/seed

test:
	cd services/auth   && go test ./...
	cd services/pet    && go test ./...
	cd services/social && go test ./...
	cd services/notif  && go test ./...
	cd services/gateway && go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w services pkg
	cd web && npm run format

tidy:
	cd services/auth    && go mod tidy
	cd services/pet     && go mod tidy
	cd services/social  && go mod tidy
	cd services/notif   && go mod tidy
	cd services/gateway && go mod tidy

web:
	cd web && npm run dev

web-install:
	cd web && npm install

web-build:
	cd web && npm run build

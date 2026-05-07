CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS pets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id      UUID NOT NULL,
    name          TEXT NOT NULL,
    species       TEXT NOT NULL DEFAULT 'blob',
    hunger        INT  NOT NULL DEFAULT 80,
    happy         INT  NOT NULL DEFAULT 80,
    energy        INT  NOT NULL DEFAULT 80,
    health        INT  NOT NULL DEFAULT 100,
    xp            INT  NOT NULL DEFAULT 0,
    stage         TEXT NOT NULL DEFAULT 'egg',
    born_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_tick_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    died_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pets_owner ON pets(owner_id);

CREATE TABLE IF NOT EXISTS pet_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pet_id      UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pet_events_pet ON pet_events(pet_id, created_at DESC);

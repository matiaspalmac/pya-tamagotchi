CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS friendships (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_a      UUID NOT NULL,
    user_b      UUID NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',  -- pending|accepted|blocked
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_a, user_b),
    CHECK (user_a <> user_b)
);

CREATE INDEX IF NOT EXISTS idx_fr_a ON friendships(user_a);
CREATE INDEX IF NOT EXISTS idx_fr_b ON friendships(user_b);

CREATE TABLE IF NOT EXISTS gifts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user    UUID NOT NULL,
    to_user      UUID NOT NULL,
    type         TEXT NOT NULL,  -- food|toy|medicine
    claimed      BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_gifts_to ON gifts(to_user, claimed, created_at DESC);

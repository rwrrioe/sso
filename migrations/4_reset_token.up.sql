BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE reset_tokens (
    token      UUID      DEFAULT gen_random_uuid() NOT NULL,
    email    VARCHAR(230)      NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '15 minutes',
    used       BOOLEAN   NOT NULL DEFAULT FALSE,

    CONSTRAINT pk_reset_tokens PRIMARY KEY (token),
);

CREATE INDEX IF NOT EXISTS idx_reset_token ON reset_tokens(token);

COMMIT;
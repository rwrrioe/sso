BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS refresh_tokens (
    token UUID DEFAULT gen_random_uuid(),
    user_id UUID,
    email VARCHAR(320),
    app_id INT,
    used BOOLEAN,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT pk_refresh_tokens PRIMARY KEY(token),
    CONSTRAINT fk_refresh_tokens_users FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_refresh_tokens_apps FOREIGN KEY (app_id) REFERENCES apps(id)
);

CREATE INDEX IF NOT EXISTS idx_refresh_token ON refresh_tokens(token);

COMMIT;
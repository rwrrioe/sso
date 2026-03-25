BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS code_types (
    id INT NOT NULL,
    type VARCHAR(20),

    CONSTRAINT pk_code_types PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS verification_codes (
    id UUID DEFAULT gen_random_uuid() NOT NULL,
    user_id UUID NOT NULL
    code VARCHAR(10) NOT NULL,
    type_id INT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_verification_codes(id),
    CONSTRAINT fk_verification_codes_users FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_verification_code_types FOREIGN KEY (type_id) REFERENCES code_types(id)
    );

CREATE INDEX IF NOT EXISTS idx_verification_code ON verification_codes(code);

COMMIT;
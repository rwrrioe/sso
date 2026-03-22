BEGIN;


CREATE TABLE IF NOT EXISTS users
(
    id UUID DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    pass_hash BYTEA NOT NULL,

    CONSTRAINT pk_users PRIMARY KEY(id)
    );
CREATE INDEX IF NOT EXISTS idx_email ON users(email);

CREATE TABLE IF NOT EXISTS apps
(
    id INT,
    name TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE,

    CONSTRAINT pk_apps PRIMARY KEY(id)
    );


CREATE TABLE IF NOT EXISTS roles
(
    id INT,
    role VARCHAR(100),

    CONSTRAINT pk_roles PRIMARY KEY(id)
    );

CREATE TABLE IF NOT EXISTS roles_users
(
    role_id INT,
    user_id UUID,

    CONSTRAINT pk_roles_users PRIMARY KEY(role_id, user_id),
    CONSTRAINT fk_roles_users_users FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_roles_users_roles FOREIGN KEY (role_id) REFERENCES roles(id)
    );

INSERT INTO apps (id, name, secret)
VALUES
    (1, 'pythia_backend', 'lieben')
    ON CONFLICT DO NOTHING;

COMMIT;
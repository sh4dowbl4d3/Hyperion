CREATE TABLE roles (
    id smallint PRIMARY KEY,
    name text NOT NULL UNIQUE
);

INSERT INTO roles (id, name) VALUES (1, 'user'), (2, 'admin');

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    password_hash text NOT NULL,
    role_id smallint NOT NULL REFERENCES roles(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_lowercase CHECK (email = lower(email)),
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE INDEX idx_users_role_id ON users (role_id);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    email_ct BYTEA NOT NULL,
    email_iv BYTEA NOT NULL,
    email_tag BYTEA NOT NULL,
    email_lookup BYTEA UNIQUE NOT NULL,
    password_hash BYTEA NOT NULL,
    password_salt BYTEA NOT NULL,
    mfa_secret_ct BYTEA,
    pin_ct BYTEA,
    pin_iv BYTEA,
    pin_tag BYTEA,
    pin_lookup BYTEA UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email_lookup ON users(email_lookup);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_pin_lookup ON users(pin_lookup) WHERE pin_lookup IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- +migrate Up
CREATE TABLE IF NOT EXISTS grants (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by TEXT NOT NULL,
    UNIQUE(username, role_id)
);

CREATE INDEX IF NOT EXISTS idx_grants_username ON grants(username);
CREATE INDEX IF NOT EXISTS idx_grants_role_id ON grants(role_id);

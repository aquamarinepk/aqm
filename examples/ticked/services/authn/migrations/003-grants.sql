-- +migrate Up
CREATE TABLE IF NOT EXISTS grants (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by TEXT NOT NULL,
    UNIQUE(user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_grants_user_id ON grants(user_id);
CREATE INDEX IF NOT EXISTS idx_grants_role_id ON grants(role_id);

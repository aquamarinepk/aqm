-- +migrate Up
-- Create todo_lists table (aggregate root)
CREATE TABLE IF NOT EXISTS todo_lists (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(user_id)
);

CREATE INDEX idx_todo_lists_user_id ON todo_lists(user_id);

-- Create todo_items table (child entities)
CREATE TABLE IF NOT EXISTS todo_items (
    id UUID PRIMARY KEY,
    list_id UUID NOT NULL REFERENCES todo_lists(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_text_length CHECK (length(text) > 0 AND length(text) <= 500)
);

CREATE INDEX idx_todo_items_list_id ON todo_items(list_id);
CREATE INDEX idx_todo_items_created_at ON todo_items(created_at);

CREATE TABLE notebook (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id UUID DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    updated_by TEXT DEFAULT NULL,
    is_deleted BOOL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    deleted_by TEXT DEFAULT NULL
);
ALTER TABLE notebook
ADD CONSTRAINT fk_notebook_parent_id FOREIGN KEY (parent_id) REFERENCES notebook(id);
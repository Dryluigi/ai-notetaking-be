CREATE TABLE "embedding_notes" (
    id UUID PRIMARY KEY,
    original_text TEXT NOT NULL,
    embedding VECTOR(768) NOT NULL,
    note_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    updated_by TEXT DEFAULT NULL,
    is_deleted BOOL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    deleted_by TEXT DEFAULT NULL
);
ALTER TABLE "embedding_notes"
ADD CONSTRAINT "fk_embedding_notes_notes" FOREIGN KEY (note_id) REFERENCES notes(id);
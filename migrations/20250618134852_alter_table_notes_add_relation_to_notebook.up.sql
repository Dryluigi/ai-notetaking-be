ALTER TABLE notes
ADD COLUMN notebook_id UUID DEFAULT NULL;
ALTER TABLE notes
ADD CONSTRAINT fk_notes_notebook_id FOREIGN KEY (notebook_id) REFERENCES notebook(id);
-- A worker waits for a free slot when the server already runs as many as it may (F10-05).
ALTER TABLE workers ADD COLUMN queued INTEGER NOT NULL DEFAULT 0;

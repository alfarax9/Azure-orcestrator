-- Tokens an agent's own transcript reports, one row per message (F14-04). A message that is
-- reported again (the transcript is read from its start after a daemon restart) replaces its row.
CREATE TABLE usage (
    worker_id   TEXT NOT NULL REFERENCES workers(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    at          INTEGER NOT NULL,
    model       TEXT NOT NULL DEFAULT '',
    input       INTEGER NOT NULL DEFAULT 0,
    output      INTEGER NOT NULL DEFAULT 0,
    cache_read  INTEGER NOT NULL DEFAULT 0,
    cache_write INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (worker_id, key)
) STRICT, WITHOUT ROWID;

CREATE INDEX usage_by_time ON usage (at);

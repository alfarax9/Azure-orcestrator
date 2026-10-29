-- Pairing tokens and the devices that exchanged them (F1-05, ADR 0006). Only SHA-256 hashes
-- of tokens are kept, so a copy of the database cannot be used to get in.
CREATE TABLE pairings (
    token_hash TEXT PRIMARY KEY,
    expires_at INTEGER NOT NULL
) STRICT;

CREATE TABLE devices (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    created_at INTEGER NOT NULL,
    last_seen  INTEGER NOT NULL
) STRICT;

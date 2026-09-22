-- Pengguna: akun login (username + password) with one Peran (Kasir/Admin).
-- Password is stored as a bcrypt hash, never plaintext. active=0 blocks login
-- without deleting history (a resigned Kasir can no longer sign in).
CREATE TABLE pengguna (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK (role IN ('kasir', 'admin')),
    active        INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

UPDATE app_meta SET value = '2' WHERE key = 'schema_version';

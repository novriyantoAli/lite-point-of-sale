-- Produk: the catalogue of what this store sells (CONTEXT.md, Produk).
--
-- `code` is the optional Kode: a barcode or an internal code used to find a
-- Produk quickly. SQLite lets a UNIQUE column hold many NULLs, which is exactly
-- the rule here — one Produk may have no Kode, but a Kode that exists is unique.
--
-- `price` and `stock` are integers: money has no decimals in this app and stock
-- is whole units per Item. A price of 0 is a free Produk; a negative one is not
-- a Produk at all, which is why both columns carry a CHECK.
--
-- `sold` records that the Produk was part of a finished Penjualan. Such a
-- Produk can only be deactivated, never deleted, so the history it belongs to
-- stays readable (CONTEXT.md, Nonaktif).
CREATE TABLE produk (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    code       TEXT UNIQUE,
    price      INTEGER NOT NULL CHECK (price >= 0),
    category   TEXT,
    stock      INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    active     INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    sold       INTEGER NOT NULL DEFAULT 0 CHECK (sold IN (0, 1)),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

UPDATE app_meta SET value = '3' WHERE key = 'schema_version';

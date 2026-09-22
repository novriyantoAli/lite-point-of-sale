-- Initial schema. The walking skeleton only needs a row the health check can
-- read to prove the database is migrated, reachable and readable; domain
-- tables (Produk, Penjualan, Stok, ...) arrive with their own slices.
CREATE TABLE app_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO app_meta (key, value) VALUES ('schema_version', '1');

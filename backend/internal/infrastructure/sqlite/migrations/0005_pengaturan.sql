-- Pengaturan: the store's one row of settings an Admin may change (CONTEXT.md,
-- Pengaturan). One store, one terminal (ADR-0002) makes a single typed row the
-- honest shape: the header/footer Struk template and the ambang Stok menipis.
--
-- `id = 1` is fixed: there is exactly one store, so there is exactly one row,
-- and the CHECK makes "a second store" impossible to write rather than a rule
-- callers have to remember. The row is seeded here, not created by Go — a read
-- never has to handle "no row yet" (ADR-0017, keputusan 3).
--
-- `paper_width` is in millimetres and only 58 or 80: those are the two thermal
-- roll widths the encoder derives its column count from (ADR-0017, keputusan 4).
--
-- `low_stock_threshold` must be above zero: an ambang of 0 would call every
-- Active Produk menipis, and a negative one is not a Stok at all. The value 5
-- is the constant this slice moves out of `domainproduk.LowStockThreshold`.
CREATE TABLE pengaturan (
    id                   INTEGER PRIMARY KEY CHECK (id = 1),
    header               TEXT NOT NULL DEFAULT '',
    footer               TEXT NOT NULL DEFAULT '',
    paper_width          INTEGER NOT NULL CHECK (paper_width IN (58, 80)),
    low_stock_threshold  INTEGER NOT NULL CHECK (low_stock_threshold > 0)
);

INSERT INTO pengaturan (id, header, footer, paper_width, low_stock_threshold)
VALUES (1, '', '', 80, 5);

UPDATE app_meta SET value = '5' WHERE key = 'schema_version';

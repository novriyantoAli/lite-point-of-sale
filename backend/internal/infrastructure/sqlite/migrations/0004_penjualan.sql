-- Penjualan: a finished sale (checkout) and the Item lines that belong to it
-- (CONTEXT.md, Penjualan, Item, Pembayaran).
--
-- `receipt_number` is the Nomor Struk: a global sequence that never resets and
-- is never reused. It is assigned inside the checkout transaction as
-- MAX(receipt_number) + 1 — one store, one terminal, one write connection
-- (ADR-0002), so the read and the write cannot interleave with another checkout.
-- The UNIQUE constraint is the backstop that makes a collision an error rather
-- than a silent duplicate Struk.
--
-- `cashier_name` is a copy of the Pengguna's username as it stood at checkout.
-- That is the name printed on the Struk, and a later rename must not rewrite a
-- Struk that was already handed over.
--
-- `total`, `amount` and `change_amount` are integers: money has no decimals in
-- this app (CONTEXT.md). For Tunai, `amount` is what the buyer handed over and
-- `change_amount` is the Kembalian; `change_amount >= 0` is the rule that a
-- payment below the total is refused. The CHECK on `method` already names all
-- four methods so #7 (non-tunai) needs no second migration.
--
-- `created_at` is store-local time (`datetime('now','localtime')`), not UTC:
-- there is one store and one terminal (ADR-0002), and both the daily revenue
-- report (#9) and the printed Struk (#8) are read in store-local time.
CREATE TABLE penjualan (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    receipt_number INTEGER NOT NULL UNIQUE,
    cashier_id     INTEGER NOT NULL REFERENCES pengguna(id),
    cashier_name   TEXT NOT NULL,
    total          INTEGER NOT NULL CHECK (total >= 0),
    method         TEXT NOT NULL CHECK (method IN ('cash', 'qris', 'debit', 'transfer')),
    amount         INTEGER NOT NULL CHECK (amount >= 0),
    change_amount  INTEGER NOT NULL DEFAULT 0 CHECK (change_amount >= 0),
    created_at     TEXT NOT NULL DEFAULT (datetime('now', 'localtime'))
);

-- Item lines. `name` and `price` are snapshots copied at checkout, so a Produk
-- renamed or repriced afterwards cannot rewrite this sale; `product_id` is kept
-- as well, so the line still says which Produk it came from. The foreign key is
-- safe because a Produk that was ever sold can only be deactivated, never
-- deleted (CONTEXT.md, Nonaktif) — checkout marks it sold in the same
-- transaction that inserts this row.
CREATE TABLE penjualan_item (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    penjualan_id INTEGER NOT NULL REFERENCES penjualan(id),
    product_id   INTEGER NOT NULL REFERENCES produk(id),
    name         TEXT NOT NULL,
    price        INTEGER NOT NULL CHECK (price >= 0),
    quantity     INTEGER NOT NULL CHECK (quantity > 0)
);

CREATE INDEX penjualan_item_penjualan ON penjualan_item (penjualan_id);

UPDATE app_meta SET value = '4' WHERE key = 'schema_version';

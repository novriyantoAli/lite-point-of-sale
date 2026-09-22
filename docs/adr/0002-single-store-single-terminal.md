# Single store, single terminal

Sistem ini dibangun untuk **satu toko dan satu terminal kasir**; schema tidak memiliki `store_id` maupun konsep terminal. Multi-toko/multi-terminal (berbagi stok antar mesin, identitas toko di tiap tabel, sinkronisasi) sengaja tidak masuk MVP.

## Considered Options

- Mendesain schema siap multi-store sejak awal — menambah `store_id` di banyak tabel dan kerumitan sinkronisasi tanpa kebutuhan nyata saat ini.

## Consequences

- Menambahkan multi-store/terminal di kemudian hari berarti migrasi schema (`store_id`) + desain sinkronisasi stok — refactor yang berarti.

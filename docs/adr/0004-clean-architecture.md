# Clean Architecture untuk backend Go

Backend Go disusun mengikuti Clean Architecture dengan dependensi yang mengarah ke dalam: `domain` (entitas & aturan domain — Produk, Penjualan, Stok), `usecase` (logika bisnis / use case), `adapter` (implementasi repository, HTTP handler, printer ESC/POS), dan `infrastructure` (SQLite, server HTTP, konfigurasi). SvelteKit dipisah antara BFF (proxy + session) dan UI. Tes tetap di seam API (perilaku eksternal), bukan detail internal lapisan.

## Considered Options

- **Struktur datar / satu package** — paling sederhana, tapi mencampur aturan domain dengan SQLite/HTTP; ditolak demi keterujian & kejelasan.
- **Layout DDD + dependency-injection framework** (gaya Fx + GORM + PostgreSQL) — terlalu berat; ditolak karena memaksa ORM/Postgres, bertentangan dengan SQLite + driver pure-Go yang ringan.

## Consequences

- Tiap slice berikutnya mengikuti struktur layer ini; `domain` tetap bebas impor SQLite/HTTP.

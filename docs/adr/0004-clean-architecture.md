# Clean Architecture untuk backend Go

Backend Go disusun mengikuti Clean Architecture dengan dependensi yang mengarah ke dalam: `domain` (entitas & aturan domain — Produk, Penjualan, Stok), `usecase` (logika bisnis / use case), `adapter` (implementasi repository, HTTP handler, printer ESC/POS), dan `infrastructure` (SQLite, server HTTP, konfigurasi). SvelteKit dipisah antara BFF (proxy + session) dan UI. Tes tetap di seam API (perilaku eksternal), bukan detail internal lapisan.

## Considered Options

- **Struktur datar / satu package** — paling sederhana, tapi mencampur aturan domain dengan SQLite/HTTP; ditolak demi keterujian & kejelasan.
- **Layout DDD + dependency-injection framework** (gaya Fx + GORM + PostgreSQL) — terlalu berat; ditolak karena memaksa ORM/Postgres, bertentangan dengan SQLite + driver pure-Go yang ringan.

## Strategi migrasi DB

Kemudahan migrasi ke DB yang lebih besar (PostgreSQL) datang dari **seam repository**, bukan dari ORM/DI framework: `domain`/`usecase` hanya bergantung pada interface repository di `adapter`, sehingga pindah dari SQLite ke Postgres cukup menulis ulang implementasi `adapter` — tanpa menyentuh aturan domain atau use case.

- PostgreSQL dijadikan **target saat multi-store/multi-terminal** benar-benar dibutuhkan (lihat ADR 0002), bukan di MVP. GORM/Fx dapat diadopsi terpisah nanti bila ukuran app menuntutnya; keduanya tidak wajib untuk migrasi DB.

## DDD: tactical dipakai, strategic ditunda

DDD di sini adalah **pendekatan pemodelan**, bukan tumpukan tool. Yang dipakai: tactical building blocks — Entity, Value Object, Aggregate, dan Repository (interface) di layer `domain`, serta `CONTEXT.md` sebagai ubiquitous language. Strategic DDD (bounded contexts lintas tim) ditunda sampai kebutuhan multi-store/multi-terminal muncul. Fx/GORM/PostgreSQL bukan bagian dari DDD — penolakannya soal tooling, bukan soal konsep DDD.

## Consequences

- Tiap slice berikutnya mengikuti struktur layer ini; `domain` tetap bebas impor SQLite/HTTP.
- Interface repository di `adapter` wajib dijaga bersih (tanpa bocor tipe SQLite ke `domain`/`usecase`) agar migrasi DB tetap murah.

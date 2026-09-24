# Lite Point of Sale

Sistem kasir ringan untuk **satu toko, satu terminal** — mencatat penjualan, mencetak struk, dan mengurangi stok.

Istilah domain ada di [`CONTEXT.md`](CONTEXT.md); keputusan arsitektur ada di [`docs/adr/`](docs/adr/).

## Struktur

```
backend/   Go — REST API bisnis + SQLite
frontend/  SvelteKit — UI + BFF (proxy ke Go, pemegang sesi)
docs/      ADR & dokumen
.github/   CI (satu workflow, dua job)
```

Dua proses (ADR-0001): browser hanya bicara ke SvelteKit; SvelteKit meneruskan panggilan ke Go; hanya Go yang membuka SQLite.

## Menjalankan

Butuh Go 1.26+, Node 22+, dan pnpm 10.

**1. API Go** (`:8080`, database `backend/data/pos.db`):

```sh
cd backend
go run ./cmd/server
```

| Env | Default | Arti |
| --- | --- | --- |
| `POS_HTTP_ADDR` | `:8080` | alamat listen API |
| `POS_DB_PATH` | `./data/pos.db` | file SQLite (migrasi dijalankan otomatis saat start) |
| `POS_TOKEN_SECRET` | `pos-development-token-secret` | kunci penanda tangan token sesi — **wajib diganti di produksi** |
| `POS_SESSION_TTL` | `12h` | masa berlaku token sesi (ADR-0010) |
| `POS_ADMIN_USERNAME` | `admin` | username Admin pertama |
| `POS_ADMIN_PASSWORD` | `admin123` | password Admin pertama (dipakai hanya saat database masih kosong) |

Saat start, store yang belum punya Admin akan **menyemai satu Admin** dari `POS_ADMIN_*` — itu satu-satunya cara membuat Pengguna pertama, karena endpoint pembuat Pengguna sendiri hanya boleh dipanggil Admin. Kalau `POS_TOKEN_SECRET` atau `POS_ADMIN_PASSWORD` masih default, server mencatat peringatan di log.

**2. UI SvelteKit** (`:5173`):

```sh
cd frontend
pnpm install
pnpm dev
```

| Env | Default | Arti |
| --- | --- | --- |
| `BACKEND_URL` | `http://localhost:8080` | base URL API Go yang diproksi BFF |
| `SESSION_MAX_AGE_SECONDS` | `43200` (12 jam) | umur cookie sesi — setidaknya sebesar `POS_SESSION_TTL` |

Buka <http://localhost:5173> — tanpa sesi kamu diarahkan ke **/login**. Login dengan `admin` / `admin123` (atau nilai `POS_ADMIN_*` yang kamu set). Setelah masuk, kartu **Status layanan** menampilkan `OK` yang berasal dari Go lewat BFF (`/api/health` → Go → SQLite). Menu **Kasir** ada untuk kedua peran — temukan Produk lewat Kode (scan/ketik) atau nama, susun keranjang, lalu bayar Tunai atau non-tunai (QRIS/Debit/Transfer); Penjualan tersimpan bersama Nomor Struk, dan Stok berkurang sendiri. Tunai menghasilkan Kembalian; non-tunai dicatat sebesar total tanpa gateway (ADR-0016). Admin punya menu **Pengguna** untuk menambah Kasir atau menonaktifkan akun, menu **Produk** untuk mengelola katalog (tambah, ubah, Nonaktifkan, atau hapus selama belum pernah terjual), serta menu **Stok** untuk mencatat barang masuk dan melihat Produk yang Stok-nya menipis atau habis.

Sesi dipegang SvelteKit sebagai cookie httpOnly berisi token internal dari Go: browser tidak pernah melihat tokennya (ADR-0001, ADR-0006, ADR-0010).

Versi produksi:

```sh
cd frontend
pnpm build
BACKEND_URL=http://localhost:8080 node build   # :3000
```

## Tes

```sh
cd backend && go vet ./... && go build ./... && go test ./...
cd frontend && pnpm check && pnpm lint && pnpm test && pnpm build
```

End-to-end (Playwright menyalakan sendiri kedua proses — butuh Go + Node terpasang):

```sh
cd frontend && pnpm test:e2e
```

- `backend/tests/e2e` memanggil **API Go lewat HTTP** (`httptest` + SQLite nyata) — lapisan e2e piramida tes (ADR-0007). Termasuk jalur autentikasi (panggilan tanpa token ditolak, peran Kasir ditolak di rute Admin, Pengguna yang dinonaktifkan kehilangan sesinya) dan jalur Penjualan (checkout atomik, blokir Stok, Kembalian, pencatatan Pembayaran Tunai maupun non-tunai, keunikan Nomor Struk, `sold` sehingga hapus Produk terjual ditolak 409).
- Tes frontend menguji schema zod, lapisan `api`, `state` (runes), proxy BFF (termasuk cookie sesi), penjaga rute `hooks.server.ts`, dan komponen lewat seam `api`/client (ADR-0007).
- `frontend/tests/e2e/*.spec.ts` menjalankan **browser sungguhan** ke build produksi: login/logout, batas peran, katalog Produk (CRUD, Kode unik, Nonaktif, saring/filter), Stok (restock aditif, penolakan jumlah nol, daftar Stok menipis beserta ambangnya), Penjualan (Tunai dengan Kembalian, non-tunai QRIS/Debit/Transfer, cari via Kode & nama, keranjang, blokir Stok, Nomor Struk, peran Kasir, Produk terjual tidak bisa dihapus), dan dashboard yang menampilkan `OK` dari Go lewat BFF (ADR-0009).

CI menjalankan semuanya pada `push`/`pull_request` ke `main` — job `backend`, `frontend`, dan `e2e` (`.github/workflows/ci.yml`, ADR-0008 + ADR-0009).

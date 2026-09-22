# Lite Point of Sale

Sistem kasir ringan untuk **satu toko, satu terminal** — mencatat penjualan, mencetak struk, dan mengurangi stok.

Istilah domain ada di [`CONTEXT.md`](CONTEXT.md); keputusan arsitektur ada di [`docs/adr/`](docs/adr/).

## Struktur

```
backend/   Go — REST API bisnis + SQLite
frontend/  SvelteKit — UI + BFF (proxy ke Go)
docs/      ADR & dokumen
.github/   CI (satu workflow, dua job)
```

Dua proses (ADR-0001): browser hanya bicara ke SvelteKit; SvelteKit meneruskan panggilan ke Go; hanya Go yang membuka SQLite.

## Menjalankan

Butuh Go 1.25+, Node 22+, dan pnpm 10.

**1. API Go** (`:8080`, database `backend/data/pos.db`):

```sh
cd backend
go run ./cmd/server
```

| Env | Default | Arti |
| --- | --- | --- |
| `POS_HTTP_ADDR` | `:8080` | alamat listen API |
| `POS_DB_PATH` | `./data/pos.db` | file SQLite (migrasi dijalankan otomatis saat start) |

**2. UI SvelteKit** (`:5173`):

```sh
cd frontend
pnpm install
pnpm dev
```

| Env | Default | Arti |
| --- | --- | --- |
| `BACKEND_URL` | `http://localhost:8080` | base URL API Go yang diproksi BFF |

Buka <http://localhost:5173> — kartu **Status layanan** menampilkan `OK` yang berasal dari Go lewat BFF (`/api/health` → Go → SQLite).

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

- `backend/tests/e2e` memanggil **API Go lewat HTTP** (`httptest` + SQLite nyata) — lapisan e2e piramida tes (ADR-0007).
- Tes frontend menguji schema zod, lapisan `api`, dan komponen lewat `api` palsu (ADR-0007).

CI menjalankan keduanya pada `push`/`pull_request` ke `main` (`.github/workflows/ci.yml`, ADR-0008).

# Monorepo: backend/ (Go) + frontend/ (SvelteKit)

Satu repo berisi dua aplikasi yang berjalan sebagai dua proses terpisah (ADR 0001): `backend/` (Go, API bisnis) dan `frontend/` (SvelteKit, UI + BFF). Tanpa tooling monorepo (turborepo/Nx/pnpm-workspace) karena belum ada package bersama.

## Considered Options

- **Polyrepo** (dua git repo terpisah) — ditolak: untuk satu tim memperlambat (PR dobel, versi antar repo, susah sinkron perubahan BFF+API).
- **Satu package campur** (frontend & backend dalam satu proses) — ditolak: bertentangan ADR 0001 (dua proses).
- **Monorepo + workspace tooling** — ditunda: belum ada package bersama yang butuh orkestrasi.

## Struktur

```
backend/   → go.mod, cmd/server/main.go, internal/{domain,usecase,adapter,infrastructure}
frontend/  → SvelteKit (lihat ADR 0006)
docs/      → ADR & dokumen
```

## Consequences

- Perubahan lintas BFF+API bisa satu commit, satu CI.
- `domain`/`usecase` wajib bebas impor SQLite/HTTP (ADR 0004) — tidak berubah.
- Monorepo tooling bisa diputuskan ulang bila muncul package bersama.

# Frontend — SvelteKit (UI + BFF)

UI aplikasi kasir sekaligus **backend-for-frontend**: browser hanya bicara ke origin ini, dan route `src/routes/api/**/+server.ts` yang meneruskan panggilan ke API Go (ADR-0001, ADR-0006).

Struktur & aturan layering ada di [ADR-0006](../docs/adr/0006-frontend-domain-slices.md); cara menjalankan kedua proses ada di [README root](../README.md).

## Perintah

```sh
pnpm dev        # dev server (:5173)
pnpm build      # build produksi (adapter-node, output: build/)
pnpm check      # svelte-check (TypeScript strict)
pnpm lint       # prettier --check + eslint
pnpm test       # vitest (unit + komponen)
pnpm test:e2e   # Playwright: build produksi + dua proses nyata (Go + SQLite)
pnpm format     # prettier --write
```

## Konfigurasi

| Env           | Default                 | Arti                                                                    |
| ------------- | ----------------------- | ----------------------------------------------------------------------- |
| `BACKEND_URL` | `http://localhost:8080` | base URL API Go yang diproksi BFF (server-only, `$env/dynamic/private`) |

Variabel client-side harus ber-prefix `PUBLIC_` dan dibaca lewat `src/lib/config/env.ts` — jangan pakai `import.meta.env` langsung di kode domain.

## Domain slice

Satu domain = satu folder di `src/lib/domains/<domain>/` berisi `schemas/` (zod, sumber tipe) → `api/` (axios + parse) → `queries/` (TanStack Query) → `components/`, dengan `index.ts` sebagai satu-satunya permukaan publik. Contoh: `src/lib/domains/health/`.

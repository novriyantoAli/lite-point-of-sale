# Frontend: vertical slice per domain + BFF proxy

Frontend SvelteKit disusun per domain (satu folder = satu konsep bisnis) dengan layering satu arah, memakai pola BFF same-origin proxy (konsisten ADR 0001).

## Tech stack

Svelte 5 (runes) + TypeScript strict · Tailwind v4 · shadcn-svelte · axios (satu instance) · TanStack Query (server state) · zod (schema = sumber tipe & validasi) · pnpm · Vitest + Playwright.

## Struktur

```
lib/domains/<domain>/{schemas,api,queries,state,components,index.ts}
lib/components/ui (shadcn, generated) + shared
lib/api/client.ts (satu axios instance, baseURL "/api")
routes/(app)/...  +  routes/api/**/+server.ts (proxy ke Go)
```

## Layering (satu arah)

`route → component → query/state → api → client`

- Component tidak import axios/fetch langsung.
- `queries/` panggil `api/` domain sendiri.
- zod schema satu-satunya yang menyeberang batas HTTP (parse di `api/`).

## BFF (Pattern A, konsisten ADR 0001)

Browser tidak pernah lihat token. `apiClient` panggil `/api/**` (same-origin, cookie httpOnly otomatis ikut); `+server.ts` baca cookie di server dan teruskan ke Go dengan Bearer token.

## Considered Options

- **Komponen per halaman tanpa domain boundary** — ditolak: logika data menyebar, susah di-test & dipindah.
- **Global store untuk server data** (Redux-style) — ditolak: duplikasi & stale; server state di TanStack Query.
- **Browser langsung ke Go** (Pattern B) — ditolak: bertentangan ADR 0001 (satu origin via BFF).

## Consequences

- Satu domain = satu vertical slice; akses lintas-domain hanya lewat `index.ts` barrel.
- Schema zod mirror 1:1 ke DTO Go backend.

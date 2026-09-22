# Strategi testing: test di seam, piramida per sisi

Prinsip: tes mengunci perilaku lewat batas interface (seam), bukan detail internal. Piramida: banyak unit, sedikit integration, beberapa e2e.

## Backend (Go)

- Unit `domain`: aturan murni (stok ≥ 0, kembalian, Penjualan final) — tanpa mock.
- Unit `usecase`: fake repository in-memory (implement interface), tanpa SQLite.
- Integration `adapter`: repository melawan SQLite nyata — `:memory:` atau file sementara di `t.TempDir()`. Tes repository memakai file sementara, karena `sqlite.Open` menyetel WAL dan satu koneksi tulis, dan dua hal itu hanya bermakna pada database berfile.
- e2e HTTP: `httptest` untuk handler.

## Frontend (SvelteKit)

- Unit schema (zod): parse fixture valid/invalid.
- Unit `api`: mock axios (axios-mock-adapter / MSW), assert shape.
- Unit `queries`: inject fake api (interface) — tanpa jaringan.
- Unit `state` (runes): class test.
- Component: `@testing-library/svelte` + `QueryClientProvider`, fresh `QueryClient` per test.
- E2E: Playwright, satu happy-path per domain.

## Aturan kunci

- Mock di level `api`/client, bukan backend Go (frontend bicara ke BFF).
- Fresh `QueryClient` tiap test (hindari cache bocor antar test).
- Interface `api/` (frontend) & repository (backend) = seam utama untuk inject fake.

## Consequences

- Kriteria selesai: `go test` + `pnpm build/check/lint/test` hijau; domain baru wajib ada test schema & api (frontend) atau unit usecase (backend).

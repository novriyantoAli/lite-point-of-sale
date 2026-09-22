# E2E Playwright di CI: satu job ketiga

Halaman sungguhan sudah ada (dashboard health), jadi syarat penundaan di ADR-0008 terpenuhi: **Playwright** kini jalan di CI sebagai job ketiga — `e2e` — di workflow yang sama. Tes menyalakan **dua proses nyata** (Go API + SQLite, dan build produksi SvelteKit lewat `node build`), lalu browser sungguhan memuat dashboard. Tidak ada mock.

## Konteks

ADR-0008 menunda "E2E Playwright di CI sejak awal" karena butuh instalasi browser dan lebih lambat, dan akan dimasukkan "saat ada halaman nyata". Sekarang ada: walking skeleton UI → BFF → Go → SQLite (ADR-0001, issue #2).

## Keputusan

- `frontend/playwright.config.ts` menyalakan kedua proses lewat `webServer` (array), masing-masing dengan database sendiri (`POS_DB_PATH=./data/e2e.db`) supaya tidak menyentuh data dev.
- `pnpm test:e2e` = `pnpm build && playwright test` — yang diuji adalah build produksi, bukan dev server.
- Job `e2e` di `.github/workflows/ci.yml` memakai **kedua toolchain** (Go + Node/pnpm): `pnpm install --frozen-lockfile` → `pnpm exec playwright install --with-deps chromium` → `pnpm test:e2e`.
- Satu happy-path per domain (ADR-0007): dashboard memuat `OK` + `Basis data: ok` dari Go.

## Considered Options

- **E2E menempel di job `frontend`** — ditolak: job itu jadi butuh Go juga, dan kegagalan e2e mengaburkan kegagalan unit/build.
- **Dev server (`pnpm dev`) sebagai target e2e** — ditolak: yang diuji bukan build yang dikirim, dan build produksi hanya butuh ~3 detik.
- **Menambah cabang e2e (database mati, dst.)** — ditolak: jalur gagal sudah dikunci oleh tes unit/komponen dan e2e Go; e2e browser cukup mengunci pipa bahagia.

## Consequences

- CI menambah satu job (unduh browser ~1–2 menit); jalur kritis UI → BFF → Go → SQLite kini terjaga otomatis, bukan cuma verifikasi manual.
- Playwright dan Vitest berbagi folder `tests/` (`tests/e2e/**` vs `tests/vitest-setup-client.ts`) pada toolchain yang sama.
- Bila e2e mulai lambat/flaky, job ini bisa dipisah, dijadwalkan ulang, atau ditambah cache browser tanpa mengubah job backend/frontend.

# E2E Playwright di CI: satu job ketiga

Halaman sungguhan sudah ada (dashboard health), jadi syarat penundaan di ADR-0008 terpenuhi: **Playwright** kini jalan di CI sebagai job ketiga — `e2e` — di workflow yang sama. Tes menyalakan **dua proses nyata** (Go API + SQLite, dan build produksi SvelteKit lewat `node build`), lalu browser sungguhan memuat dashboard. Tidak ada mock.

## Konteks

ADR-0008 menunda "E2E Playwright di CI sejak awal" karena butuh instalasi browser dan lebih lambat, dan akan dimasukkan "saat ada halaman nyata". Sekarang ada: walking skeleton UI → BFF → Go → SQLite (ADR-0001, issue #2).

## Keputusan

- `frontend/playwright.config.ts` menyalakan kedua proses lewat `webServer` (array), masing-masing dengan database sendiri (`POS_DB_PATH=./data/e2e.db`) supaya tidak menyentuh data dev.
- Karena slice autentikasi (#3) semua halaman butuh sesi, konfigurasi itu juga menetapkan `POS_TOKEN_SECRET` dan kredensial Admin seed (`POS_ADMIN_*`) secara eksplisit — suite tidak bergantung pada default pengembangan.
- Store e2e **unik per run** (`backend/data/e2e-<pid>.db`, dihapus `global-teardown.ts` sesudahnya): Go hanya menyemai Admin pada store kosong, jadi file sisa dari run lama bisa memuat kredensial yang tidak lagi dikenal suite.
- `reuseExistingServer: false` untuk kedua proses. Playwright menyalakan `webServer` **sebelum** global setup, sehingga menghapus database dari setup hook justru menghapus file yang sudah dipegang server — suite lalu berjalan di store run sebelumnya. Server Go juga membawa state, dan server SvelteKit adalah build: memakai ulang keduanya berarti menguji sesuatu selain yang baru dibangun, jadi port yang sudah terpakai lebih baik gagal terang-terangan.
- `pnpm test:e2e` = `pnpm build && playwright test` — yang diuji adalah build produksi, bukan dev server.
- Job `e2e` di `.github/workflows/ci.yml` memakai **kedua toolchain** (Go + Node/pnpm): `pnpm install --frozen-lockfile` → `pnpm exec playwright install --with-deps chromium` → `pnpm test:e2e`.
- Satu happy-path per domain (ADR-0007): dashboard memuat `OK` + `Basis data: ok` dari Go.
- Halaman yang butuh sesi membuat spec-nya login dulu lewat formulir sungguhan (`tests/e2e/helpers.ts`), bukan dengan menyuntik cookie — jalur login itu sendiri bagian dari yang diuji.

## Considered Options

- **E2E menempel di job `frontend`** — ditolak: job itu jadi butuh Go juga, dan kegagalan e2e mengaburkan kegagalan unit/build.
- **Dev server (`pnpm dev`) sebagai target e2e** — ditolak: yang diuji bukan build yang dikirim, dan build produksi hanya butuh ~3 detik.
- **Menambah cabang e2e (database mati, dst.)** — ditolak: jalur gagal sudah dikunci oleh tes unit/komponen dan e2e Go; e2e browser cukup mengunci pipa bahagia.

## Consequences

- CI menambah satu job (unduh browser ~1–2 menit); jalur kritis UI → BFF → Go → SQLite kini terjaga otomatis, bukan cuma verifikasi manual.
- Playwright dan Vitest berbagi folder `tests/` (`tests/e2e/**` vs `tests/vitest-setup-client.ts`) pada toolchain yang sama.
- Bila e2e mulai lambat/flaky, job ini bisa dipisah, dijadwalkan ulang, atau ditambah cache browser tanpa mengubah job backend/frontend.

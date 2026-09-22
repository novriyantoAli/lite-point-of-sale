# CI: GitHub Actions, satu workflow dua job

CI memakai **GitHub Actions** dengan satu workflow `.github/workflows/ci.yml` berisi dua job sesuai monorepo (ADR 0005): `backend` (Go) dan `frontend` (SvelteKit). Trigger `push` + `pull_request` ke `main`.

## Job backend

`go vet ./...` → `go build ./...` → `go test ./...` (working-directory `backend/`).

## Job frontend

`pnpm install --frozen-lockfile` → `pnpm check` → `pnpm lint` → `pnpm test` → `pnpm build` (working-directory `frontend/`).

## Considered Options

- **CI terpisah per sisi / multi-workflow** — ditunda: satu workflow cukup, lebih mudah dibaca; dipecah bila sudah besar.
- **E2E Playwright di CI sejak awal** — ditunda: butuh instalasi browser & lebih lambat; dimasukkan saat ada halaman nyata (fase lanjut ADR 0007).

## Consequences

- Setiap issue build harus menyertakan konfigurasi CI-nya ke workflow ini, tidak membuat workflow baru ad hoc.
- Kriteria selesai: semua job hijau di `push`/`pull_request`.

## Agent skills

### Issue tracker

GitHub Issues (repo `novriyantoAli/lite-point-of-sale`); external PRs are **not** a triage surface. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical triage roles map to GitHub labels with default names: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context — `CONTEXT.md` at the repo root plus `docs/adr/`. See `docs/agents/domain.md`.

### Documented solutions

`docs/solutions/` — catatan solusi masalah yang sudah lewat (bug, praktik, pola alur kerja), diorganisir per kategori dengan YAML frontmatter (`module`, `tags`, `problem_type`). Relevan saat mengimplementasi atau men-debug di area yang sudah terdokumentasi.

### Code organization

Backend Go follows Clean Architecture (`domain` → `usecase` → `adapter` → `infrastructure`, dependencies point inward); SvelteKit keeps the BFF separate from the UI. Monorepo: `backend/` (Go) + `frontend/` (SvelteKit). See `docs/adr/0004-clean-architecture.md`, `0005-monorepo-repo-structure.md`, `0006-frontend-domain-slices.md`, `0007-testing-strategy.md`.

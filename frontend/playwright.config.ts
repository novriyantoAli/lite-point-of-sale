import { fileURLToPath } from 'node:url';
import { defineConfig, devices } from '@playwright/test';

/** Absolute paths, so nothing depends on a relative walk from somewhere else. */
const BACKEND_DIR = fileURLToPath(new URL('../backend', import.meta.url));
/**
 * The E2E store, unique per run.
 *
 * Playwright starts `webServer` *before* global setup, so a database wiped from
 * a setup hook is one the Go API already has open: the server keeps working on
 * the unlinked file, and the suite silently runs against the previous run's
 * store. A path that cannot exist yet removes the race entirely, and
 * global-teardown.ts deletes it afterwards.
 */
const E2E_DB_PATH = fileURLToPath(
	new URL(`../backend/data/e2e-${process.pid}.db`, import.meta.url)
);

const SVELTEKIT_PORT = 3000;
const GO_API_PORT = 8080;

/**
 * Credentials of the seeded Admin Pengguna. Passed to the Go API explicitly
 * rather than relying on its development defaults, and used by the login specs
 * through tests/e2e/helpers.ts.
 */
const E2E_ADMIN_USERNAME = 'admin';
const E2E_ADMIN_PASSWORD = 'rahasia-admin';
const E2E_TOKEN_SECRET = 'rahasia-uji-e2e';

/**
 * E2E runs against the production build of the two real processes (ADR-0009):
 * the Go API on SQLite plus the SvelteKit server (UI + BFF). Playwright starts
 * both, on a database of their own, so the suite never touches dev data.
 *
 * Neither server is ever reused, not even locally: the Go API carries the store
 * the suite writes to, and the SvelteKit server is a build — reusing either
 * means testing something other than what was just built. An occupied port now
 * fails loudly instead of quietly testing yesterday's processes.
 */
export default defineConfig({
	testDir: './tests/e2e',
	// The store this run uses; global-teardown.ts reads it from here — one
	// source of truth for a path both files need.
	metadata: { e2eDbPath: E2E_DB_PATH },
	globalTeardown: './tests/e2e/global-teardown.ts',
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 1 : 0,
	reporter: process.env.CI ? [['github'], ['list']] : 'list',
	use: {
		baseURL: `http://127.0.0.1:${SVELTEKIT_PORT}`,
		trace: 'on-first-retry'
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
	webServer: [
		{
			command: 'go run ./cmd/server',
			cwd: BACKEND_DIR,
			env: {
				POS_HTTP_ADDR: `:${GO_API_PORT}`,
				POS_DB_PATH: E2E_DB_PATH,
				POS_TOKEN_SECRET: E2E_TOKEN_SECRET,
				POS_ADMIN_USERNAME: E2E_ADMIN_USERNAME,
				POS_ADMIN_PASSWORD: E2E_ADMIN_PASSWORD
			},
			url: `http://127.0.0.1:${GO_API_PORT}/api/health`,
			reuseExistingServer: false,
			timeout: 120_000
		},
		{
			command: 'node build',
			env: {
				BACKEND_URL: `http://127.0.0.1:${GO_API_PORT}`,
				PORT: String(SVELTEKIT_PORT),
				ORIGIN: `http://127.0.0.1:${SVELTEKIT_PORT}`
			},
			url: `http://127.0.0.1:${SVELTEKIT_PORT}`,
			reuseExistingServer: false,
			timeout: 120_000
		}
	]
});

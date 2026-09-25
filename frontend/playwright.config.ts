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
/**
 * The file that stands in for the thermal printer of the Go process.
 *
 * `POS_PRINTER_DEVICE` points at it, so a checkout really writes ESC/POS bytes
 * through the device adapter — the success path of a print runs in the browser
 * suite instead of only ever the failure (ADR-0017). global-setup.ts creates it
 * and global-teardown.ts removes it, both from this one metadata entry: the
 * adapter opens a device path and deliberately never creates one, because a typo
 * in POS_PRINTER_DEVICE has to fail loudly rather than print into a new file.
 */
const E2E_PRINTER_PATH = fileURLToPath(
	new URL(`../backend/data/e2e-printer-${process.pid}.bin`, import.meta.url)
);
/**
 * The folder the Go API writes its automatic and manual backup snapshots into.
 * Pointing it here keeps a browser run from writing into the repo's
 * ./data/backup — the suite gets a folder of its own, removed by
 * global-teardown.ts. The adapter creates it on its first snapshot, so there is
 * no setup hook for it.
 */
const E2E_BACKUP_DIR = fileURLToPath(
	new URL(`../backend/data/e2e-backup-${process.pid}`, import.meta.url)
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
	// The store this run uses and the printer file it prints to; global-teardown.ts
	// reads both from here — one source of truth for paths two files need.
	metadata: { e2eDbPath: E2E_DB_PATH, e2ePrinterPath: E2E_PRINTER_PATH, e2eBackupDir: E2E_BACKUP_DIR },
	globalSetup: './tests/e2e/global-setup.ts',
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
				POS_ADMIN_PASSWORD: E2E_ADMIN_PASSWORD,
				// The printer the suite prints to: the file above, not a device.
				POS_PRINTER_DEVICE: E2E_PRINTER_PATH,
				// The backup folder the suite snapshots into: a run of its own,
				// not the repo's default.
				POS_BACKUP_DIR: E2E_BACKUP_DIR
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

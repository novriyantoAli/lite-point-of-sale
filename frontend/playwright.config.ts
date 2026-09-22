import { defineConfig, devices } from '@playwright/test';

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
 * both, each on its own database file, so the suite never touches dev data.
 */
export default defineConfig({
	testDir: './tests/e2e',
	// A store left over from an earlier run could hold an Admin whose password
	// this suite no longer knows, so every run starts from an empty database.
	globalSetup: './tests/e2e/global-setup.ts',
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
			cwd: '../backend',
			env: {
				POS_HTTP_ADDR: `:${GO_API_PORT}`,
				POS_DB_PATH: './data/e2e.db',
				POS_TOKEN_SECRET: E2E_TOKEN_SECRET,
				POS_ADMIN_USERNAME: E2E_ADMIN_USERNAME,
				POS_ADMIN_PASSWORD: E2E_ADMIN_PASSWORD
			},
			url: `http://127.0.0.1:${GO_API_PORT}/api/health`,
			reuseExistingServer: !process.env.CI,
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
			reuseExistingServer: !process.env.CI,
			timeout: 120_000
		}
	]
});

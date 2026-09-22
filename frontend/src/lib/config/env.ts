import { env as privateEnv } from '$env/dynamic/private';

const DEFAULT_BACKEND_URL = 'http://localhost:8080';

/** Matches the default of POS_SESSION_TTL on the Go side. */
const DEFAULT_SESSION_MAX_AGE_SECONDS = 12 * 60 * 60;

/**
 * Server-only configuration, read at runtime (`$env/dynamic/private`).
 *
 * Import this from server code only — `+server.ts` routes and server load
 * functions. Never from a component or any module the browser can reach:
 * SvelteKit refuses to bundle private env into client code by design.
 */
export const serverEnv = {
	/** Base URL of the Go API the BFF proxies to. */
	get backendUrl(): string {
		return privateEnv.BACKEND_URL || DEFAULT_BACKEND_URL;
	},

	/**
	 * Lifetime of the BFF's session cookie, in seconds.
	 *
	 * Set it to at least the Go API's POS_SESSION_TTL. The two are separate
	 * processes, so neither can read the other's configuration: a cookie that
	 * dies first logs the terminal out early, while one that outlives its token
	 * only costs a trip back to the login page.
	 */
	get sessionMaxAgeSeconds(): number {
		const configured = Number(privateEnv.SESSION_MAX_AGE_SECONDS);
		if (!Number.isFinite(configured) || configured <= 0) {
			return DEFAULT_SESSION_MAX_AGE_SECONDS;
		}

		return configured;
	}
};

import { env as privateEnv } from '$env/dynamic/private';

const DEFAULT_BACKEND_URL = 'http://localhost:8080';

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
	}
};

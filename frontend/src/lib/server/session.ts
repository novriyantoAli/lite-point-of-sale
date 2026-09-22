import { serverEnv } from '$lib/config/env';

/**
 * Server-only plumbing for the session cookie.
 *
 * The cookie holds the Go token and nothing else: it is httpOnly, so no script
 * on the page can read it, and it is never written by the browser — only by the
 * BFF login route (ADR-0001, ADR-0006).
 */

/** Name of the cookie carrying the Go token. */
export const SESSION_COOKIE = 'session';

/**
 * Cookie attributes. `secure` follows the protocol actually in use: a store
 * running this on `http://localhost` is the MVP deployment (ADR-0002), and a
 * hard-coded `secure: true` would silently refuse to set the cookie there,
 * while an HTTPS deployment still gets the flag.
 */
export function sessionCookieOptions(url: URL) {
	return {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		secure: url.protocol === 'https:',
		maxAge: serverEnv.sessionMaxAgeSeconds
	} as const;
}

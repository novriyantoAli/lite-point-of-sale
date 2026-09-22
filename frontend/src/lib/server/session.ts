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
 * How long the cookie lives, kept in step with the default of POS_SESSION_TTL
 * on the Go side. A mismatch is survivable in both directions: a cookie that
 * outlives its token gets a 401 from Go and a trip back to the login page, and
 * one that dies early costs a login.
 */
const SESSION_MAX_AGE_SECONDS = 12 * 60 * 60;

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
		maxAge: SESSION_MAX_AGE_SECONDS
	} as const;
}

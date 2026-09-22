import { SESSION_COOKIE, sessionCookieOptions } from '$lib/server/session';
import type { RequestHandler } from './$types';

/**
 * Logout is dropping the cookie — there is no Go call to make.
 *
 * The Go token is stateless: it carries no server-side record, so there is
 * nothing to revoke, and it simply stops being presented. The trade-off is
 * deliberate and written down in ADR-0010: a token that has already leaked
 * stays valid until it expires.
 */
export const POST: RequestHandler = ({ cookies, url }) => {
	cookies.delete(SESSION_COOKIE, sessionCookieOptions(url));

	return new Response(null, { status: 204 });
};

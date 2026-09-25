import { redirect, type Handle } from '@sveltejs/kit';
import { readSession } from '$lib/server/backend';

/**
 * Routes only an Admin may open. Hiding the menu item is not a guard — a Kasir
 * can type a URL — so the check runs here, on the server, before anything is
 * rendered.
 */
const ADMIN_ONLY = ['/produk', '/stok', '/pengguna', '/pengaturan', '/laporan', '/backup'];

const LOGIN_ROUTE = '/login';

/**
 * The session guard of the app (ADR-0001, ADR-0006). The session lives in an
 * httpOnly cookie, so only the server can answer "who is this?" — every page
 * navigation asks Go once and puts the answer in `event.locals`.
 */
export const handle: Handle = async ({ event, resolve }) => {
	// The BFF's own routes are not pages. They answer the API and let Go decide
	// (401/403) instead of being redirected to a login form, and they must not
	// pay for a session lookup the route itself is about to make.
	if (event.url.pathname.startsWith('/api/')) {
		return resolve(event);
	}

	// Anything that is not a page — files in static/, an unknown URL — has no
	// session to resolve and no page to guard. Deciding on the route instead of
	// on an Accept header is what keeps this a guard rather than a suggestion:
	// a client that asks for JSON without a session still gets turned away.
	if (!event.route.id) {
		return resolve(event);
	}

	const user = await readSession(event);
	event.locals.user = user;

	const isLoginRoute = event.url.pathname === LOGIN_ROUTE;

	if (!user) {
		return isLoginRoute ? resolve(event) : redirect(303, LOGIN_ROUTE);
	}

	// Somebody is already logged in: the login form has nothing to offer.
	if (isLoginRoute) {
		return redirect(303, '/');
	}

	if (!ADMIN_ONLY.some((route) => event.url.pathname.startsWith(route)) || user.role === 'admin') {
		return resolve(event);
	}

	// A Kasir who lands on an Admin page goes back to the till. The action
	// behind it would have been refused by Go anyway (403).
	return redirect(303, '/');
};

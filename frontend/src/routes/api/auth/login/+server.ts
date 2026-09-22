import { json } from '@sveltejs/kit';
import { BackendSessionSchema } from '$lib/domains/auth';
import { callBackend, forward } from '$lib/server/backend';
import { SESSION_COOKIE, sessionCookieOptions } from '$lib/server/session';
import type { RequestHandler } from './$types';

/**
 * Login for the browser, and the only place a session cookie is handed out.
 *
 * This is Pattern A (ADR-0001, ADR-0006): the route calls Go, keeps the token
 * Go answers with inside an httpOnly cookie, and tells the browser only who it
 * logged in as. Client-side JavaScript never sees the token — that is what
 * httpOnly is for, and it is why no interceptor here ever builds an
 * Authorization header.
 */
export const POST: RequestHandler = async (event) => {
	const response = await callBackend(event, '/api/auth/login', {
		method: 'POST',
		body: await event.request.text()
	});

	// Wrong credentials or an unusable body: Go's status and envelope go
	// straight through, so the form shows the message Go wrote.
	if (!response.ok) {
		return forward(response);
	}

	const session = BackendSessionSchema.parse(await response.json());
	event.cookies.set(SESSION_COOKIE, session.data.token, sessionCookieOptions(event.url));

	return json({ data: { user: session.data.user } });
};

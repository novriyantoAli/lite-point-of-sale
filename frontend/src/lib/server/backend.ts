import { error, json, type RequestEvent } from '@sveltejs/kit';
import { serverEnv } from '$lib/config/env';
import { SessionSchema, type Pengguna } from '$lib/domains/auth';
import { SESSION_COOKIE } from './session';

/**
 * Server-only helpers for talking to the Go API. They live in `$lib/server`,
 * which SvelteKit refuses to bundle into the browser — the same reason the
 * backend URL lives here and not in a component.
 */

/** The app's error envelope for a Go API that cannot be reached. */
export const BACKEND_UNREACHABLE = {
	message: 'Tidak dapat menghubungi server.',
	error: 'backend_unreachable'
} as const;

/** What the BFF needs to call Go on the browser's behalf. */
export type BackendCaller = Pick<RequestEvent, 'fetch' | 'cookies'>;

/**
 * Forwards a Go response to the browser: same status, same body, same content
 * type. Errors keep the shape Go gave them, so `lib/api/errors.ts` stays the
 * only place that turns a failure into a message.
 */
export function forward(response: Response): Response {
	return new Response(response.body, {
		status: response.status,
		headers: {
			'content-type': response.headers.get('content-type') ?? 'application/json'
		}
	});
}

/**
 * Calls the Go API for the browser, attaching the session token it cannot see:
 * the token lives in an httpOnly cookie this server-side code reads (ADR-0001,
 * ADR-0006). A Go API that cannot be reached answers with the app's error
 * envelope rather than letting a generic SvelteKit 500 through.
 */
export async function callBackend(
	event: BackendCaller,
	path: string,
	options: { method?: string; body?: string } = {}
): Promise<Response> {
	const token = event.cookies.get(SESSION_COOKIE);

	const headers: Record<string, string> = {};
	if (token) {
		headers.authorization = `Bearer ${token}`;
	}
	if (options.body !== undefined) {
		headers['content-type'] = 'application/json';
	}

	try {
		return await event.fetch(`${serverEnv.backendUrl}${path}`, {
			method: options.method ?? 'GET',
			headers,
			body: options.body
		});
	} catch {
		return json(BACKEND_UNREACHABLE, { status: 502 });
	}
}

/**
 * Who the session cookie belongs to, or null when there is no session.
 *
 * Go is asked on the server for every guarded request instead of the cookie
 * carrying a copy of the Pengguna: a deactivated account then loses access on
 * its next navigation rather than whenever a cookie happens to expire.
 *
 * A 401 is an answer ("nobody is logged in"), so it returns null. A Go API
 * that cannot be reached is not, so it fails loudly instead of looking like a
 * logout.
 */
export async function readSession(event: BackendCaller): Promise<Pengguna | null> {
	if (!event.cookies.get(SESSION_COOKIE)) {
		return null;
	}

	const response = await callBackend(event, '/api/auth/me');
	if (response.status === 401) {
		return null;
	}
	if (!response.ok) {
		throw error(502, 'Tidak dapat memeriksa sesi ke API.');
	}

	return SessionSchema.parse(await response.json()).data.user;
}

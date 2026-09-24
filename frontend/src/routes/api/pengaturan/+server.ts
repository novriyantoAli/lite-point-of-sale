import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * The store's Pengaturan, proxied to Go with the session token the browser
 * cannot see. The Peran guard lives in Go: this route does not decide who may
 * call it, it just forwards, and a Kasir gets Go's 403 back.
 *
 * The body is passed through untouched — the paper width and the ambang are
 * Go's to validate, and parsing them here would be a second opinion about what
 * they mean.
 */
export const GET: RequestHandler = async (event) =>
	forward(await callBackend(event, '/api/pengaturan', { method: 'GET' }));

export const PUT: RequestHandler = async (event) =>
	forward(
		await callBackend(event, '/api/pengaturan', {
			method: 'PUT',
			body: await event.request.text()
		})
	);

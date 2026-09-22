import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * Staff list and staff creation, proxied to Go with the session token the
 * browser cannot see. The Peran guard lives in Go: this route does not decide
 * who may call it, it just forwards, and a Kasir gets Go's 403 back.
 */
export const GET: RequestHandler = async (event) =>
	forward(await callBackend(event, '/api/pengguna', { method: 'GET' }));

export const POST: RequestHandler = async (event) =>
	forward(
		await callBackend(event, '/api/pengguna', {
			method: 'POST',
			body: await event.request.text()
		})
	);

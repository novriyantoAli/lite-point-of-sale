import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * The catalogue, proxied to Go with the session token the browser cannot see.
 * The Peran guard lives in Go: this route does not decide who may call it, it
 * just forwards, and a Kasir gets Go's 403 back.
 *
 * The query string is passed through untouched — the filters are Go's to read,
 * and re-encoding them here would be a second opinion about what they mean.
 */
export const GET: RequestHandler = async (event) =>
	forward(await callBackend(event, `/api/produk${event.url.search}`, { method: 'GET' }));

export const POST: RequestHandler = async (event) =>
	forward(
		await callBackend(event, '/api/produk', {
			method: 'POST',
			body: await event.request.text()
		})
	);

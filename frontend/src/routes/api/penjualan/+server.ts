import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * Checking a keranjang out into a Penjualan, proxied to Go with the session token
 * the browser cannot see.
 *
 * There is no Peran guard here and none in Go: both Peran sell at a one-terminal
 * store, so this is the one part of the API behind the token check alone.
 *
 * The body is passed through untouched — the cart is Go's to price and validate,
 * and parsing it here would be a second opinion about what it means.
 */
export const POST: RequestHandler = async (event) =>
	forward(
		await callBackend(event, '/api/penjualan', {
			method: 'POST',
			body: await event.request.text()
		})
	);

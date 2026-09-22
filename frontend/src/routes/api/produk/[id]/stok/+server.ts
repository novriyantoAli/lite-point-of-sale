import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * Recording a restock, proxied to Go with the session token the browser cannot
 * see. The Peran guard lives in Go: this route does not decide who may call it,
 * it just forwards, and a Kasir gets Go's 403 back.
 *
 * The body is passed through untouched — the quantity is Go's to read and
 * validate, and parsing it here would be a second opinion about what it means.
 */
export const POST: RequestHandler = async (event) =>
	forward(
		await callBackend(event, `/api/produk/${event.params.id}/stok`, {
			method: 'POST',
			body: await event.request.text()
		})
	);

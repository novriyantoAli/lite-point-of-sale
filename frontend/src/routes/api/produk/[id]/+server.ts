import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * Changing, deactivating, and deleting one Produk, proxied to Go. The three
 * verbs are separate on purpose: an edit replaces the editable fields, while
 * Active is changed by its own PATCH so an edit form cannot rewrite it by
 * accident (see `adapter/httpapi/produk.go`).
 *
 * Whether a delete is allowed at all — a Produk that already sold may only be
 * deactivated — is the use case's call in Go, and it answers 409.
 */
export const PUT: RequestHandler = async (event) =>
	forward(
		await callBackend(event, `/api/produk/${event.params.id}`, {
			method: 'PUT',
			body: await event.request.text()
		})
	);

export const PATCH: RequestHandler = async (event) =>
	forward(
		await callBackend(event, `/api/produk/${event.params.id}`, {
			method: 'PATCH',
			body: await event.request.text()
		})
	);

export const DELETE: RequestHandler = async (event) =>
	forward(await callBackend(event, `/api/produk/${event.params.id}`, { method: 'DELETE' }));

import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * Activating or deactivating a Pengguna, proxied to Go. Who the acting Admin is
 * comes from the token, not from this request: the API refuses to let an Admin
 * deactivate their own account, and it has to be Go that decides that.
 */
export const PATCH: RequestHandler = async (event) =>
	forward(
		await callBackend(event, `/api/pengguna/${event.params.id}`, {
			method: 'PATCH',
			body: await event.request.text()
		})
	);

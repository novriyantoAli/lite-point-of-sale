import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * The Kategori already in use, proxied to Go. This route is a literal segment
 * next to the `[id]` one, so SvelteKit prefers it — the same way Go's mux
 * registers it before the wildcard. "Kategori" is not a Produk id.
 */
export const GET: RequestHandler = async (event) =>
	forward(await callBackend(event, '/api/produk/kategori', { method: 'GET' }));

import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * The omzet of one store-local day (#9), proxied to Go with the session token the
 * browser cannot see.
 *
 * It is Admin-only in Go, and this route adds no guard of its own: a Kasir's
 * request keeps Go's 403, which is the same answer the page guard gives by
 * redirecting them away from `/laporan`.
 *
 * The query string is passed through untouched — the day is Go's to validate, and
 * parsing it here would be a second opinion about what a report date means.
 */
export const GET: RequestHandler = async (event) =>
	forward(await callBackend(event, `/api/penjualan/omzet${event.url.search}`, { method: 'GET' }));

import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * The store's backups, proxied to Go with the session token the browser cannot
 * see. The Peran guard lives in Go: this route does not decide who may call it,
 * it just forwards, and a Kasir gets Go's 403 back.
 *
 * There is no body to pass through: a backup is "snapshot now", and the server
 * decides what that means.
 */
export const GET: RequestHandler = async (event) =>
	forward(await callBackend(event, '/api/backup', { method: 'GET' }));

export const POST: RequestHandler = async (event) =>
	forward(await callBackend(event, '/api/backup', { method: 'POST' }));

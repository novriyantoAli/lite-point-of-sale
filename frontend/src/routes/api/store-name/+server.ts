import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * BFF proxy for the store's name — the one value the browser reads before login.
 * Go serves `/api/store-name` without a guard (ADR-0019), so a session cookie is
 * neither required nor attached here; the name, and only the name, comes back.
 */
export const GET: RequestHandler = async (event) =>
	forward(await callBackend(event, '/api/store-name', { method: 'GET' }));

import { serverEnv } from '$lib/config/env';
import type { RequestHandler } from './$types';

/**
 * BFF proxy for the health endpoint. The browser calls `/api/health` on this
 * origin; the server forwards it to the Go API, so no backend URL or token is
 * ever exposed to the client (ADR-0001, ADR-0006).
 */
export const GET: RequestHandler = async ({ fetch }) => {
	const response = await fetch(`${serverEnv.backendUrl}/api/health`);

	return new Response(response.body, {
		status: response.status,
		headers: {
			'content-type': response.headers.get('content-type') ?? 'application/json'
		}
	});
};

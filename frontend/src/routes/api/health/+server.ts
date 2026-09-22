import { json } from '@sveltejs/kit';
import { serverEnv } from '$lib/config/env';
import type { RequestHandler } from './$types';

/** The app's error envelope: `{ message, error }`, normalized by `lib/api/errors`. */
const BACKEND_UNREACHABLE = {
	message: 'Tidak dapat menghubungi server.',
	error: 'backend_unreachable'
} as const;

/**
 * BFF proxy for the health endpoint. The browser calls `/api/health` on this
 * origin; the server forwards it to the Go API, so no backend URL or token is
 * ever exposed to the client (ADR-0001, ADR-0006).
 */
export const GET: RequestHandler = async ({ fetch }) => {
	try {
		const response = await fetch(`${serverEnv.backendUrl}/api/health`);

		return new Response(response.body, {
			status: response.status,
			headers: {
				'content-type': response.headers.get('content-type') ?? 'application/json'
			}
		});
	} catch {
		// This is the only layer that knows Go is unreachable — answer with the
		// app's error envelope instead of letting SvelteKit's generic 500 through.
		return json(BACKEND_UNREACHABLE, { status: 502 });
	}
};

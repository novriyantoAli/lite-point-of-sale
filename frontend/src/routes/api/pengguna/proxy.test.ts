import { describe, expect, it, vi } from 'vitest';
import { PATCH } from './[id]/+server';
import { GET, POST } from './+server';

vi.mock('$lib/config/env', () => ({
	serverEnv: { backendUrl: 'http://backend.test', sessionMaxAgeSeconds: 43_200 }
}));

/** A request arriving at the BFF with a session cookie already set. */
function eventWith<T>(
	fetch: typeof globalThis.fetch,
	options: { method: string; body?: string } = { method: 'GET' }
): T {
	const url = 'http://localhost/api/pengguna/7';

	return {
		fetch,
		url: new URL(url),
		params: { id: '7' },
		request: new Request(url, { method: options.method, body: options.body }),
		cookies: { get: () => 'token-dari-cookie', set: vi.fn(), delete: vi.fn() }
	} as T;
}

describe('/api/pengguna (BFF proxy)', () => {
	it('forwards the list request to Go with the session token the browser cannot read', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: [] }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			})
		);

		const response = await GET(eventWith<Parameters<typeof GET>[0]>(fetch));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/pengguna',
			expect.objectContaining({
				method: 'GET',
				headers: expect.objectContaining({ authorization: 'Bearer token-dari-cookie' })
			})
		);
		expect(response.status).toBe(200);
		expect(await response.json()).toEqual({ data: [] });
	});

	it('lets the Peran guard stay in Go: a Kasir gets the 403 back untouched', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(
				JSON.stringify({
					message: 'Anda tidak berhak melakukan tindakan ini.',
					error: 'forbidden'
				}),
				{ status: 403, headers: { 'content-type': 'application/json' } }
			)
		);

		const response = await GET(eventWith<Parameters<typeof GET>[0]>(fetch));

		expect(response.status).toBe(403);
		expect(await response.json()).toEqual({
			message: 'Anda tidak berhak melakukan tindakan ini.',
			error: 'forbidden'
		});
	});

	it('forwards the body of a new Pengguna', async () => {
		const body = JSON.stringify({ username: 'kasir1', password: 'rahasia123', role: 'kasir' });
		const fetch = vi.fn().mockResolvedValue(new Response('{"data":{"user":{}}}', { status: 201 }));

		await POST(eventWith<Parameters<typeof POST>[0]>(fetch, { method: 'POST', body }));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/pengguna',
			expect.objectContaining({ method: 'POST', body })
		);
	});

	it('forwards a deactivation to the Pengguna the path names', async () => {
		const body = JSON.stringify({ active: false });
		const fetch = vi.fn().mockResolvedValue(new Response('{"data":{"user":{}}}', { status: 200 }));

		await PATCH(eventWith<Parameters<typeof PATCH>[0]>(fetch, { method: 'PATCH', body }));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/pengguna/7',
			expect.objectContaining({ method: 'PATCH', body })
		);
	});
});

import { describe, expect, it, vi } from 'vitest';
import { GET, PUT } from './+server';

vi.mock('$lib/config/env', () => ({
	serverEnv: { backendUrl: 'http://backend.test', sessionMaxAgeSeconds: 43_200 }
}));

/** A request arriving at the BFF with a session cookie already set. */
function eventWith(fetch: typeof globalThis.fetch, body?: string) {
	const url = 'http://localhost/api/pengaturan';

	return {
		fetch,
		url: new URL(url),
		request: new Request(url, { method: body === undefined ? 'GET' : 'PUT', body }),
		cookies: { get: () => 'token-dari-cookie', set: vi.fn(), delete: vi.fn() }
	} as unknown as Parameters<typeof GET>[0];
}

describe('/api/pengaturan (BFF proxy)', () => {
	it('forwards the read to Go with the session token the browser cannot read', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { pengaturan: { id: 1, paper_width: 80 } } }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			})
		);

		const response = await GET(eventWith(fetch));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/pengaturan',
			expect.objectContaining({
				method: 'GET',
				headers: expect.objectContaining({ authorization: 'Bearer token-dari-cookie' })
			})
		);
		expect(response.status).toBe(200);
		expect(await response.json()).toEqual({ data: { pengaturan: { id: 1, paper_width: 80 } } });
	});

	it('forwards the new Pengaturan untouched: the rules are Go\u2019s', async () => {
		const body = JSON.stringify({
			header: 'Toko Kopi',
			footer: 'Terima kasih',
			paper_width: 58,
			low_stock_threshold: 3
		});
		const fetch = vi.fn().mockResolvedValue(new Response('{}', { status: 200 }));

		await PUT(eventWith(fetch, body));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/pengaturan',
			expect.objectContaining({ method: 'PUT', body })
		);
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

		const response = await GET(eventWith(fetch));

		expect(response.status).toBe(403);
		expect(await response.json()).toEqual({
			message: 'Anda tidak berhak melakukan tindakan ini.',
			error: 'forbidden'
		});
	});

	it('answers the app error envelope when Go is unreachable', async () => {
		const fetch = vi.fn().mockRejectedValue(new TypeError('fetch failed'));

		const response = await GET(eventWith(fetch));

		expect(response.status).toBe(502);
		expect(await response.json()).toEqual({
			message: 'Tidak dapat menghubungi server.',
			error: 'backend_unreachable'
		});
	});
});

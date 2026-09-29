import { describe, expect, it, vi } from 'vitest';
import { GET } from './+server';

vi.mock('$lib/config/env', () => ({
	serverEnv: { backendUrl: 'http://backend.test', sessionMaxAgeSeconds: 43_200 }
}));

/** A request arriving at the BFF before any session exists. */
function eventWith(fetch: typeof globalThis.fetch) {
	const url = 'http://localhost/api/store-name';

	return {
		fetch,
		url: new URL(url),
		request: new Request(url, { method: 'GET' }),
		cookies: { get: () => undefined }
	} as unknown as Parameters<typeof GET>[0];
}

describe('/api/store-name (BFF proxy)', () => {
	it('forwards the public read and sends no session token', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { store_name: 'Toko Kopi' } }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			})
		);

		const response = await GET(eventWith(fetch));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/store-name',
			expect.objectContaining({
				method: 'GET',
				headers: expect.not.objectContaining({ authorization: expect.anything() })
			})
		);
		expect(response.status).toBe(200);
		expect(await response.json()).toEqual({ data: { store_name: 'Toko Kopi' } });
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

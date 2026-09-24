import { describe, expect, it, vi } from 'vitest';
import { GET } from './+server';

vi.mock('$lib/config/env', () => ({
	serverEnv: { backendUrl: 'http://backend.test', sessionMaxAgeSeconds: 43_200 }
}));

/** A lookup arriving at the BFF, with a session cookie unless `token` is null. */
function eventWith(
	fetch: typeof globalThis.fetch,
	receiptNumber = '7',
	token: string | null = 'token-dari-cookie'
) {
	const url = `http://localhost/api/penjualan/${receiptNumber}`;

	return {
		fetch,
		url: new URL(url),
		params: { receiptNumber },
		request: new Request(url, { method: 'GET' }),
		cookies: { get: () => token, set: vi.fn(), delete: vi.fn() }
	} as unknown as Parameters<typeof GET>[0];
}

const sale = {
	receipt_number: 7,
	created_at: '2026-09-23 10:00:00',
	cashier_id: 1,
	cashier_name: 'kasir1',
	total: 36000,
	items: [{ product_id: 1, name: 'Kopi Susu', price: 18000, quantity: 2, subtotal: 36000 }],
	payment: { method: 'cash', amount: 50000, change: 14000 }
};

describe('/api/penjualan/[receiptNumber] (BFF proxy)', () => {
	it('forwards the lookup to Go with the session token the browser cannot read', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { sale } }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			})
		);

		const response = await GET(eventWith(fetch));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/penjualan/7',
			expect.objectContaining({
				method: 'GET',
				headers: expect.objectContaining({ authorization: 'Bearer token-dari-cookie' })
			})
		);
		// A read sends no body at all: the Nomor Struk is the whole of the request.
		expect(fetch.mock.calls[0]![1].body).toBeUndefined();
		expect(response.status).toBe(200);
		expect(await response.json()).toEqual({ data: { sale } });
	});

	it('keeps a Nomor Struk that names nothing as the readable 404 Go gave it', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(
				JSON.stringify({ message: 'Penjualan tidak ditemukan.', error: 'sale_not_found' }),
				{
					status: 404,
					headers: { 'content-type': 'application/json' }
				}
			)
		);

		const response = await GET(eventWith(fetch, '999'));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/penjualan/999',
			expect.objectContaining({ method: 'GET' })
		);
		expect(response.status).toBe(404);
		expect(await response.json()).toEqual({
			message: 'Penjualan tidak ditemukan.',
			error: 'sale_not_found'
		});
	});

	it('sends no token when there is no session, and passes Go\u2019s 401 back', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ message: 'Sesi tidak valid.', error: 'invalid_token' }), {
				status: 401,
				headers: { 'content-type': 'application/json' }
			})
		);

		const response = await GET(eventWith(fetch, '7', null));

		const headers = fetch.mock.calls[0]![1].headers as Record<string, string>;
		expect(headers.authorization).toBeUndefined();
		expect(response.status).toBe(401);
		expect(await response.json()).toEqual({ message: 'Sesi tidak valid.', error: 'invalid_token' });
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

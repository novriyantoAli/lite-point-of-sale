import { describe, expect, it, vi } from 'vitest';
import { POST } from './+server';

vi.mock('$lib/config/env', () => ({
	serverEnv: { backendUrl: 'http://backend.test', sessionMaxAgeSeconds: 43_200 }
}));

/** A checkout arriving at the BFF with a session cookie already set. */
function eventWith(fetch: typeof globalThis.fetch) {
	const url = 'http://localhost/api/penjualan';

	return {
		fetch,
		url: new URL(url),
		request: new Request(url, {
			method: 'POST',
			body: JSON.stringify({
				items: [{ product_id: 1, quantity: 2 }],
				payment: { method: 'cash', amount: 50000 }
			})
		}),
		cookies: { get: () => 'token-dari-cookie', set: vi.fn(), delete: vi.fn() }
	} as unknown as Parameters<typeof POST>[0];
}

describe('/api/penjualan (BFF proxy)', () => {
	it('forwards the checkout to Go with the session token the browser cannot read', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { sale: { receipt_number: 1 } } }), {
				status: 201,
				headers: { 'content-type': 'application/json' }
			})
		);

		const response = await POST(eventWith(fetch));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/penjualan',
			expect.objectContaining({
				method: 'POST',
				headers: expect.objectContaining({
					authorization: 'Bearer token-dari-cookie',
					'content-type': 'application/json'
				})
			})
		);
		expect(response.status).toBe(201);
		expect(await response.json()).toEqual({ data: { sale: { receipt_number: 1 } } });
	});

	it("passes the cart through untouched: the pricing and the Stok check are Go's", async () => {
		const fetch = vi.fn().mockResolvedValue(new Response('{}', { status: 201 }));

		await POST(eventWith(fetch));

		const sent = JSON.parse(fetch.mock.calls[0]![1].body as string);
		expect(sent).toEqual({
			items: [{ product_id: 1, quantity: 2 }],
			payment: { method: 'cash', amount: 50000 }
		});
	});

	it('lets a refused checkout keep the status and the message Go gave it', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(
				JSON.stringify({
					message: 'Stok Kopi tidak cukup: tersisa 2, diminta 3.',
					error: 'insufficient_stock'
				}),
				{ status: 409, headers: { 'content-type': 'application/json' } }
			)
		);

		const response = await POST(eventWith(fetch));

		expect(response.status).toBe(409);
		expect(await response.json()).toEqual({
			message: 'Stok Kopi tidak cukup: tersisa 2, diminta 3.',
			error: 'insufficient_stock'
		});
	});

	it('answers the app error envelope when Go is unreachable', async () => {
		const fetch = vi.fn().mockRejectedValue(new TypeError('fetch failed'));

		const response = await POST(eventWith(fetch));

		expect(response.status).toBe(502);
		expect(await response.json()).toEqual({
			message: 'Tidak dapat menghubungi server.',
			error: 'backend_unreachable'
		});
	});
});

import { describe, expect, it, vi } from 'vitest';
import { POST } from './+server';

vi.mock('$lib/config/env', () => ({
	serverEnv: { backendUrl: 'http://backend.test', sessionMaxAgeSeconds: 43_200 }
}));

/** A reprint arriving at the BFF, with a session cookie unless `token` is null. */
function eventWith(
	fetch: typeof globalThis.fetch,
	receiptNumber = '7',
	token: string | null = 'token-dari-cookie'
) {
	const url = `http://localhost/api/penjualan/${receiptNumber}/struk`;

	return {
		fetch,
		url: new URL(url),
		params: { receiptNumber },
		request: new Request(url, { method: 'POST' }),
		cookies: { get: () => token, set: vi.fn(), delete: vi.fn() }
	} as unknown as Parameters<typeof POST>[0];
}

describe('/api/penjualan/[receiptNumber]/struk (BFF proxy)', () => {
	it('forwards the reprint to Go with the session token the browser cannot read', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { print: { printed: true } } }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			})
		);

		const response = await POST(eventWith(fetch));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/penjualan/7/struk',
			expect.objectContaining({
				method: 'POST',
				headers: expect.objectContaining({ authorization: 'Bearer token-dari-cookie' })
			})
		);
		// A reprint carries no body: the Nomor Struk in the path is the whole request.
		expect(fetch.mock.calls[0]![1].body).toBeUndefined();
		expect(response.status).toBe(200);
		expect(await response.json()).toEqual({ data: { print: { printed: true } } });
	});

	it('keeps a print that failed as a readable result, not an error status', async () => {
		// The Penjualan is stored either way, so a printer that did not answer is a
		// reported result the till can retry (ADR-0017, keputusan 1).
		const fetch = vi.fn().mockResolvedValue(
			new Response(
				JSON.stringify({ data: { print: { printed: false, message: 'Printer belum diatur.' } } }),
				{
					status: 200,
					headers: { 'content-type': 'application/json' }
				}
			)
		);

		const response = await POST(eventWith(fetch));

		expect(response.status).toBe(200);
		expect(await response.json()).toEqual({
			data: { print: { printed: false, message: 'Printer belum diatur.' } }
		});
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

		const response = await POST(eventWith(fetch, '999'));

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/penjualan/999/struk',
			expect.objectContaining({ method: 'POST' })
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

		const response = await POST(eventWith(fetch, '7', null));

		const headers = fetch.mock.calls[0]![1].headers as Record<string, string>;
		expect(headers.authorization).toBeUndefined();
		expect(response.status).toBe(401);
		expect(await response.json()).toEqual({ message: 'Sesi tidak valid.', error: 'invalid_token' });
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

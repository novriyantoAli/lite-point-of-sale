import { describe, expect, it, vi } from 'vitest';
import { GET } from './+server';

vi.mock('$lib/config/env', () => ({
	serverEnv: { backendUrl: 'http://backend.test' }
}));

function eventWith(fetch: typeof globalThis.fetch) {
	return { fetch } as unknown as Parameters<typeof GET>[0];
}

describe('GET /api/health (BFF proxy)', () => {
	it('forwards the Go response untouched', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response('{"status":"ok","database":"ok"}', {
				status: 200,
				headers: { 'content-type': 'application/json' }
			})
		);

		const response = await GET(eventWith(fetch));

		expect(fetch).toHaveBeenCalledWith('http://backend.test/api/health');
		expect(response.status).toBe(200);
		expect(await response.json()).toEqual({ status: 'ok', database: 'ok' });
	});

	it('keeps the Go status when the backend reports a problem', async () => {
		const fetch = vi
			.fn()
			.mockResolvedValue(
				new Response('{"status":"degraded","database":"unavailable"}', { status: 503 })
			);

		const response = await GET(eventWith(fetch));

		expect(response.status).toBe(503);
	});

	it('answers with the app error envelope when Go is unreachable', async () => {
		const fetch = vi.fn().mockRejectedValue(new TypeError('fetch failed'));

		const response = await GET(eventWith(fetch));

		expect(response.status).toBe(502);
		expect(await response.json()).toEqual({
			message: 'Tidak dapat menghubungi server.',
			error: 'backend_unreachable'
		});
	});
});

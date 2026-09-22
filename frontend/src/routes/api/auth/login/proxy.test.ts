import { describe, expect, it, vi } from 'vitest';
import { POST } from './+server';

vi.mock('$lib/config/env', () => ({
	serverEnv: { backendUrl: 'http://backend.test', sessionMaxAgeSeconds: 43_200 }
}));

const pengguna = { id: 1, username: 'admin', role: 'admin', active: true };

/** The SvelteKit event a login request arrives with, plus the cookie spy. */
function loginEvent(fetch: typeof globalThis.fetch) {
	const set = vi.fn();

	const event = {
		fetch,
		url: new URL('http://localhost/api/auth/login'),
		request: new Request('http://localhost/api/auth/login', {
			method: 'POST',
			body: JSON.stringify({ username: 'admin', password: 'rahasia123' })
		}),
		cookies: { get: vi.fn(), set, delete: vi.fn() }
	};

	return { event: event as unknown as Parameters<typeof POST>[0], set };
}

describe('POST /api/auth/login (BFF proxy)', () => {
	it('keeps the Go token in an httpOnly cookie and never returns it', async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { token: 'token-dari-go', user: pengguna } }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			})
		);
		const { event, set } = loginEvent(fetch);

		const response = await POST(event);
		const body = await response.text();

		expect(fetch).toHaveBeenCalledWith(
			'http://backend.test/api/auth/login',
			expect.objectContaining({ method: 'POST' })
		);
		expect(JSON.parse(body)).toEqual({ data: { user: pengguna } });
		expect(body).not.toContain('token-dari-go');
		expect(set).toHaveBeenCalledWith(
			'session',
			'token-dari-go',
			expect.objectContaining({ httpOnly: true, path: '/', sameSite: 'lax', maxAge: 43_200 })
		);
	});

	it("forwards Go's rejection and sets no cookie", async () => {
		const fetch = vi.fn().mockResolvedValue(
			new Response(
				JSON.stringify({
					message: 'Username atau password salah.',
					error: 'invalid_credentials'
				}),
				{ status: 401, headers: { 'content-type': 'application/json' } }
			)
		);
		const { event, set } = loginEvent(fetch);

		const response = await POST(event);

		expect(response.status).toBe(401);
		expect(await response.json()).toEqual({
			message: 'Username atau password salah.',
			error: 'invalid_credentials'
		});
		expect(set).not.toHaveBeenCalled();
	});

	it('answers the app error envelope when Go is unreachable', async () => {
		const fetch = vi.fn().mockRejectedValue(new TypeError('fetch failed'));
		const { event, set } = loginEvent(fetch);

		const response = await POST(event);

		expect(response.status).toBe(502);
		expect(await response.json()).toEqual({
			message: 'Tidak dapat menghubungi server.',
			error: 'backend_unreachable'
		});
		expect(set).not.toHaveBeenCalled();
	});

	it('fails loudly when Go answers without a token', async () => {
		const fetch = vi
			.fn()
			.mockResolvedValue(
				new Response(JSON.stringify({ data: { user: pengguna } }), { status: 200 })
			);
		const { event } = loginEvent(fetch);

		await expect(POST(event)).rejects.toThrow();
	});
});

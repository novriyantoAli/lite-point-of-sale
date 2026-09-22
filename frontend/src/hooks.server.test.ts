import { describe, expect, it, vi } from 'vitest';
import { handle } from './hooks.server';

const { readSession } = vi.hoisted(() => ({ readSession: vi.fn() }));

// The guard is tested at its seam: what it decides for a given session, without
// a Go API behind it (ADR-0007).
vi.mock('$lib/server/backend', () => ({ readSession }));

const admin = { id: 1, username: 'admin', role: 'admin' as const, active: true };
const kasir = { id: 2, username: 'kasir1', role: 'kasir' as const, active: true };

/** One navigation, as hooks.server sees it. */
function navigation(pathname: string, options: { accept?: string; dataRequest?: boolean } = {}) {
	const url = new URL(`http://localhost${pathname}`);

	const event = {
		url,
		request: new Request(url, { headers: { accept: options.accept ?? 'text/html' } }),
		isDataRequest: options.dataRequest ?? false,
		locals: {} as App.Locals,
		cookies: { get: vi.fn(), set: vi.fn(), delete: vi.fn() },
		fetch: vi.fn()
	};

	return { event: event as unknown as Parameters<typeof handle>[0]['event'], resolve: vi.fn() };
}

describe('session guard', () => {
	it('sends an anonymous visitor to the login page', async () => {
		readSession.mockResolvedValue(null);
		const { event, resolve } = navigation('/');

		await expect(handle({ event, resolve })).rejects.toMatchObject({
			status: 303,
			location: '/login'
		});
		expect(resolve).not.toHaveBeenCalled();
	});

	it('lets an anonymous visitor reach the login page', async () => {
		readSession.mockResolvedValue(null);
		const { event, resolve } = navigation('/login');

		await handle({ event, resolve });

		expect(resolve).toHaveBeenCalledOnce();
		expect(event.locals.user).toBeNull();
	});

	it('keeps a logged-in Pengguna away from the login page', async () => {
		readSession.mockResolvedValue(kasir);
		const { event, resolve } = navigation('/login');

		await expect(handle({ event, resolve })).rejects.toMatchObject({ location: '/' });
	});

	it('hands the resolved Pengguna to the rest of the request', async () => {
		readSession.mockResolvedValue(kasir);
		const { event, resolve } = navigation('/');

		await handle({ event, resolve });

		expect(resolve).toHaveBeenCalledOnce();
		expect(event.locals.user).toEqual(kasir);
	});

	it('sends a Kasir away from an Admin-only page', async () => {
		readSession.mockResolvedValue(kasir);
		const { event, resolve } = navigation('/pengguna');

		await expect(handle({ event, resolve })).rejects.toMatchObject({
			status: 303,
			location: '/'
		});
		expect(resolve).not.toHaveBeenCalled();
	});

	it('lets an Admin open an Admin-only page', async () => {
		readSession.mockResolvedValue(admin);
		const { event, resolve } = navigation('/pengguna');

		await handle({ event, resolve });

		expect(resolve).toHaveBeenCalledOnce();
	});

	it('does not resolve a session for the BFF routes: Go answers those', async () => {
		readSession.mockClear();
		const { event, resolve } = navigation('/api/pengguna');

		await handle({ event, resolve });

		expect(resolve).toHaveBeenCalledOnce();
		expect(readSession).not.toHaveBeenCalled();
	});

	it('does not resolve a session for an asset, so a dead API cannot take the UI down', async () => {
		readSession.mockClear();
		const { event, resolve } = navigation('/favicon.svg', { accept: 'image/svg+xml' });

		await handle({ event, resolve });

		expect(resolve).toHaveBeenCalledOnce();
		expect(readSession).not.toHaveBeenCalled();
	});

	it('guards a client-side navigation too, not just a full page load', async () => {
		readSession.mockResolvedValue(kasir);
		const { event, resolve } = navigation('/pengguna', {
			accept: 'application/json',
			dataRequest: true
		});

		await expect(handle({ event, resolve })).rejects.toMatchObject({ location: '/' });
	});
});

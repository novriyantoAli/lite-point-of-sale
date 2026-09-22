import MockAdapter from 'axios-mock-adapter';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { apiClient } from '$lib/api/client';
import { authApi } from './auth.api';

let mock: MockAdapter;

const pengguna = { id: 1, username: 'kasir1', role: 'kasir', active: true };

beforeEach(() => {
	mock = new MockAdapter(apiClient);
});

afterEach(() => {
	mock.restore();
});

describe('authApi.login', () => {
	it('posts the credentials to the BFF and returns the Pengguna', async () => {
		mock.onPost('/auth/login').reply(200, { data: { user: pengguna } });

		const user = await authApi.login({ username: 'kasir1', password: 'rahasia123' });

		expect(user).toEqual(pengguna);
		expect(JSON.parse(mock.history.post[0]!.data as string)).toEqual({
			username: 'kasir1',
			password: 'rahasia123'
		});
	});

	it('surfaces wrong credentials as a normalized AppError', async () => {
		mock.onPost('/auth/login').reply(401, {
			message: 'Username atau password salah.',
			error: 'invalid_credentials'
		});

		await expect(authApi.login({ username: 'admin', password: 'salah' })).rejects.toMatchObject({
			message: 'Username atau password salah.',
			status: 401,
			code: 'invalid_credentials'
		});
	});

	it('fails loudly when the answer carries no Pengguna', async () => {
		mock.onPost('/auth/login').reply(200, { data: {} });

		await expect(authApi.login({ username: 'admin', password: 'rahasia123' })).rejects.toThrow();
	});
});

describe('authApi.logout', () => {
	it('asks the BFF to drop the session cookie', async () => {
		mock.onPost('/auth/logout').reply(204);

		await expect(authApi.logout()).resolves.toBeUndefined();
		expect(mock.history.post).toHaveLength(1);
	});
});

describe('authApi.listPengguna', () => {
	it('returns the staff list', async () => {
		mock.onGet('/pengguna').reply(200, { data: [pengguna] });

		await expect(authApi.listPengguna()).resolves.toEqual([pengguna]);
	});

	it('surfaces a refused Peran as a normalized AppError', async () => {
		mock.onGet('/pengguna').reply(403, {
			message: 'Anda tidak berhak melakukan tindakan ini.',
			error: 'forbidden'
		});

		await expect(authApi.listPengguna()).rejects.toMatchObject({
			status: 403,
			code: 'forbidden'
		});
	});
});

describe('authApi.createPengguna', () => {
	it('posts a new Pengguna and returns the stored one', async () => {
		mock.onPost('/pengguna').reply(201, { data: { user: pengguna } });

		const created = await authApi.createPengguna({
			username: 'kasir1',
			password: 'rahasia123',
			role: 'kasir'
		});

		expect(created).toEqual(pengguna);
	});

	it('rejects an input the schema refuses before any request is made', async () => {
		await expect(
			authApi.createPengguna({ username: 'kasir1', password: 'pendek', role: 'kasir' })
		).rejects.toThrow();
		expect(mock.history.post).toHaveLength(0);
	});
});

describe('authApi.setPenggunaActive', () => {
	it('patches the Pengguna and returns its new state', async () => {
		mock.onPatch('/pengguna/1').reply(200, { data: { user: { ...pengguna, active: false } } });

		const updated = await authApi.setPenggunaActive(1, false);

		expect(updated.active).toBe(false);
		expect(JSON.parse(mock.history.patch[0]!.data as string)).toEqual({ active: false });
	});
});

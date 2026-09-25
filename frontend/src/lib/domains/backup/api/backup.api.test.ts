import MockAdapter from 'axios-mock-adapter';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { apiClient } from '$lib/api/client';
import { backupApi } from './backup.api';

let mock: MockAdapter;

const backup = {
	name: 'pos-20260115-120000-000000000.db',
	size: 8192,
	created_at: '2026-01-15T12:00:00Z'
};

beforeEach(() => {
	mock = new MockAdapter(apiClient);
});

afterEach(() => {
	mock.restore();
});

describe('backupApi.list', () => {
	it('returns the snapshots on disk', async () => {
		mock.onGet('/backup').reply(200, { data: [backup] });

		await expect(backupApi.list()).resolves.toEqual([backup]);
	});

	it('reads an empty list as no backups', async () => {
		mock.onGet('/backup').reply(200, { data: [] });

		await expect(backupApi.list()).resolves.toEqual([]);
	});

	it('surfaces a refused Peran as a normalized AppError', async () => {
		mock.onGet('/backup').reply(403, {
			message: 'Anda tidak berhak melakukan tindakan ini.',
			error: 'forbidden'
		});

		await expect(backupApi.list()).rejects.toMatchObject({ status: 403, code: 'forbidden' });
	});

	it('fails loudly when the answer is not a list of snapshots', async () => {
		mock.onGet('/backup').reply(200, { data: { backup } });

		await expect(backupApi.list()).rejects.toThrow();
	});
});

describe('backupApi.create', () => {
	it('takes a snapshot and returns it', async () => {
		mock.onPost('/backup').reply(201, { data: { backup } });

		await expect(backupApi.create()).resolves.toEqual(backup);
	});

	it('surfaces a server error as a normalized AppError', async () => {
		mock.onPost('/backup').reply(500, {
			message: 'Terjadi kesalahan pada server.',
			error: 'internal_error'
		});

		await expect(backupApi.create()).rejects.toMatchObject({ status: 500, code: 'internal_error' });
	});

	it('fails loudly when the answer is not a backup', async () => {
		mock.onPost('/backup').reply(201, { data: { name: 42 } });

		await expect(backupApi.create()).rejects.toThrow();
	});
});

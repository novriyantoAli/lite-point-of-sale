import MockAdapter from 'axios-mock-adapter';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { apiClient } from '$lib/api/client';
import { healthApi } from './health.api';

let mock: MockAdapter;

beforeEach(() => {
	mock = new MockAdapter(apiClient);
});

afterEach(() => {
	mock.restore();
});

describe('healthApi', () => {
	it('reads the health endpoint through the shared client and returns the parsed contract', async () => {
		mock.onGet('/health').reply(200, { status: 'ok', database: 'ok', uptimeSeconds: 12 });

		const health = await healthApi.check();

		expect(health).toEqual({ status: 'ok', database: 'ok' });
		expect(mock.history.get).toHaveLength(1);
	});

	it('fails loudly when the response does not match the contract', async () => {
		mock.onGet('/health').reply(200, { status: 'ok' });

		await expect(healthApi.check()).rejects.toThrow();
	});

	it('surfaces a failing response as a normalized AppError', async () => {
		mock.onGet('/health').reply(503, {
			message: 'Basis data tidak tersedia.',
			error: 'database_unavailable'
		});

		await expect(healthApi.check()).rejects.toMatchObject({
			message: 'Basis data tidak tersedia.',
			status: 503,
			code: 'database_unavailable'
		});
	});
});

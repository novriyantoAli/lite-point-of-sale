import { beforeEach, describe, expect, it, vi } from 'vitest';

const { get } = vi.hoisted(() => ({ get: vi.fn() }));

vi.mock('$lib/api/client', () => ({ apiClient: { get } }));

import { healthApi } from './health.api';

describe('healthApi', () => {
	beforeEach(() => {
		get.mockReset();
	});

	it('asks the BFF for the health endpoint and returns the parsed contract', async () => {
		// The BFF is addressed relative to the shared client's `/api` baseURL.
		get.mockResolvedValue({ data: { status: 'ok', database: 'ok', uptimeSeconds: 12 } });

		const health = await healthApi.check();

		expect(get).toHaveBeenCalledWith('/health');
		expect(health).toEqual({ status: 'ok', database: 'ok' });
	});

	it('fails loudly when the response does not match the contract', async () => {
		get.mockResolvedValue({ data: { status: 'ok' } });

		await expect(healthApi.check()).rejects.toThrow();
	});
});

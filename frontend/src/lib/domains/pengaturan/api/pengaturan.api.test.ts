import MockAdapter from 'axios-mock-adapter';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { apiClient } from '$lib/api/client';
import { pengaturanApi } from './pengaturan.api';

let mock: MockAdapter;

const pengaturan = {
	id: 1,
	header: 'Toko Kopi',
	footer: 'Terima kasih',
	paper_width: 80,
	low_stock_threshold: 5
};

beforeEach(() => {
	mock = new MockAdapter(apiClient);
});

afterEach(() => {
	mock.restore();
});

describe('pengaturanApi.get', () => {
	it('returns the stored Pengaturan', async () => {
		mock.onGet('/pengaturan').reply(200, { data: { settings: pengaturan } });

		await expect(pengaturanApi.get()).resolves.toEqual(pengaturan);
	});

	it('surfaces a refused Peran as a normalized AppError', async () => {
		mock.onGet('/pengaturan').reply(403, {
			message: 'Anda tidak berhak melakukan tindakan ini.',
			error: 'forbidden'
		});

		await expect(pengaturanApi.get()).rejects.toMatchObject({ status: 403, code: 'forbidden' });
	});

	it('fails loudly when the answer is not a Pengaturan', async () => {
		mock.onGet('/pengaturan').reply(200, { data: { settings: { paper_width: '80' } } });

		await expect(pengaturanApi.get()).rejects.toThrow();
	});
});

describe('pengaturanApi.storeName', () => {
	it('returns the store name the public endpoint answers', async () => {
		mock.onGet('/store-name').reply(200, { data: { store_name: 'Toko Kopi' } });

		await expect(pengaturanApi.storeName()).resolves.toBe('Toko Kopi');
	});

	it('fails loudly when the answer is not a store name', async () => {
		mock.onGet('/store-name').reply(200, { data: { store_name: 42 } });

		await expect(pengaturanApi.storeName()).rejects.toThrow();
	});
});

describe('pengaturanApi.update', () => {
	it('puts the new Pengaturan and returns the stored one', async () => {
		mock.onPut('/pengaturan').reply(200, {
			data: { settings: { ...pengaturan, paper_width: 58, low_stock_threshold: 3 } }
		});

		const updated = await pengaturanApi.update({
			header: 'Toko Kopi',
			footer: 'Terima kasih',
			paper_width: 58,
			low_stock_threshold: 3
		});

		expect(updated.paper_width).toBe(58);
		expect(updated.low_stock_threshold).toBe(3);
		expect(JSON.parse(mock.history.put[0]!.data as string)).toEqual({
			header: 'Toko Kopi',
			footer: 'Terima kasih',
			paper_width: 58,
			low_stock_threshold: 3
		});
	});

	it('rejects an input the schema refuses before any request is made', async () => {
		// `paper_width` is typed 58 | 80, so a runtime caller sending another value is
		// the case being proven here — the cast is the bypass that exercises it.
		await expect(
			pengaturanApi.update({
				header: '',
				footer: '',
				paper_width: 60 as never,
				low_stock_threshold: 5
			})
		).rejects.toThrow();
		expect(mock.history.put).toHaveLength(0);
	});

	it('surfaces a refused paper width from the API as a normalized AppError', async () => {
		mock.onPut('/pengaturan').reply(400, {
			message: 'Lebar kertas harus 58 atau 80 mm.',
			error: 'invalid_input'
		});

		await expect(
			pengaturanApi.update({ header: '', footer: '', paper_width: 80, low_stock_threshold: 5 })
		).rejects.toMatchObject({ status: 400, code: 'invalid_input' });
	});
});

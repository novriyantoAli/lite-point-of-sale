import MockAdapter from 'axios-mock-adapter';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { apiClient } from '$lib/api/client';
import { penjualanApi } from './penjualan.api';

let mock: MockAdapter;

const sale = {
	receipt_number: 1,
	created_at: '2026-09-23 10:00:00',
	cashier_id: 1,
	cashier_name: 'kasir1',
	total: 36000,
	items: [{ product_id: 1, name: 'Kopi Susu', price: 18000, quantity: 2, subtotal: 36000 }],
	payment: { method: 'cash', amount: 50000, change: 14000 }
};

/** A cart of two Kopi Susu, paid with 50000. */
const cart = {
	items: [{ product_id: 1, quantity: 2 }],
	payment: { method: 'cash' as const, amount: 50000 }
};

beforeEach(() => {
	mock = new MockAdapter(apiClient);
});

afterEach(() => {
	mock.restore();
});

describe('penjualanApi.checkout', () => {
	it('posts the cart and answers the Penjualan the API stored', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale } });

		await expect(penjualanApi.checkout(cart)).resolves.toEqual(sale);
		expect(JSON.parse(mock.history.post[0]!.data)).toEqual(cart);
	});

	it('sends only the Produk id and the quantity — never a name or a price', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale } });

		await penjualanApi.checkout(cart);

		const body = JSON.parse(mock.history.post[0]!.data);
		expect(Object.keys(body.items[0]).sort()).toEqual(['product_id', 'quantity']);
		// The Kembalian is the API's to work out, so it is never sent either.
		expect(Object.keys(body.payment).sort()).toEqual(['amount', 'method']);
	});

	it('refuses an empty keranjang before the request is sent', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale } });

		await expect(
			penjualanApi.checkout({ items: [], payment: { method: 'cash', amount: 10000 } })
		).rejects.toThrow();
		expect(mock.history.post).toHaveLength(0);
	});

	it('refuses a method other than Tunai before the request is sent', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale } });

		await expect(
			penjualanApi.checkout({
				items: [{ product_id: 1, quantity: 1 }],
				payment: { method: 'qris' as 'cash', amount: 18000 }
			})
		).rejects.toThrow();
		expect(mock.history.post).toHaveLength(0);
	});

	it('surfaces a refused checkout as a normalized AppError, message and all', async () => {
		mock.onPost('/penjualan').reply(409, {
			message: 'Stok Kopi tidak cukup: tersisa 2, diminta 3.',
			error: 'insufficient_stock'
		});

		await expect(penjualanApi.checkout(cart)).rejects.toMatchObject({
			status: 409,
			code: 'insufficient_stock',
			message: 'Stok Kopi tidak cukup: tersisa 2, diminta 3.'
		});
	});

	it('fails loudly when the answer is not a Penjualan', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale: { receipt_number: 'satu' } } });

		await expect(penjualanApi.checkout(cart)).rejects.toThrow();
	});
});

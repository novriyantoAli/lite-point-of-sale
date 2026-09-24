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

/** The print result of a Struk that came out. */
const printed = { printed: true };

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
	it('posts the cart and answers the Penjualan the API stored, print result and all', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale, print: printed } });

		await expect(penjualanApi.checkout(cart)).resolves.toEqual({ sale, print: printed });
		expect(JSON.parse(mock.history.post[0]!.data)).toEqual(cart);
	});

	it('reads a print that failed rather than treating it as a failed checkout', async () => {
		// The Penjualan is stored either way; the failure is reported, not thrown
		// (ADR-0017, keputusan 1).
		const gagal = { printed: false, message: 'Printer belum diatur.' };
		mock.onPost('/penjualan').reply(201, { data: { sale, print: gagal } });

		await expect(penjualanApi.checkout(cart)).resolves.toEqual({ sale, print: gagal });
	});

	it('sends only the Produk id and the quantity — never a name or a price', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale, print: printed } });

		await penjualanApi.checkout(cart);

		const body = JSON.parse(mock.history.post[0]!.data);
		expect(Object.keys(body.items[0]).sort()).toEqual(['product_id', 'quantity']);
		// The Kembalian is the API's to work out, so it is never sent either.
		expect(Object.keys(body.payment).sort()).toEqual(['amount', 'method']);
	});

	it('refuses an empty keranjang before the request is sent', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale, print: printed } });

		await expect(
			penjualanApi.checkout({ items: [], payment: { method: 'cash', amount: 10000 } })
		).rejects.toThrow();
		expect(mock.history.post).toHaveLength(0);
	});

	it('posts a recorded method with the total as its nominal', async () => {
		const qris = { ...sale, payment: { method: 'qris', amount: 36000, change: 0 } };
		mock.onPost('/penjualan').reply(201, { data: { sale: qris, print: printed } });

		// No gateway is involved: the method and the nominal are all that is sent,
		// and no Kembalian is declared (CONTEXT.md, Pembayaran).
		await expect(
			penjualanApi.checkout({
				items: [{ product_id: 1, quantity: 2 }],
				payment: { method: 'qris', amount: 36000 }
			})
		).resolves.toEqual({ sale: qris, print: printed });

		expect(JSON.parse(mock.history.post[0]!.data).payment).toEqual({
			method: 'qris',
			amount: 36000
		});
	});

	it('refuses a method the API does not know before the request is sent', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale, print: printed } });

		await expect(
			penjualanApi.checkout({
				items: [{ product_id: 1, quantity: 1 }],
				payment: { method: 'bitcoin' as 'cash', amount: 18000 }
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
		mock.onPost('/penjualan').reply(201, {
			data: { sale: { receipt_number: 'satu' }, print: printed }
		});

		await expect(penjualanApi.checkout(cart)).rejects.toThrow();
	});

	it('fails loudly when the answer does not say whether the Struk printed', async () => {
		mock.onPost('/penjualan').reply(201, { data: { sale } });

		await expect(penjualanApi.checkout(cart)).rejects.toThrow();
	});
});

describe('penjualanApi.getByReceiptNumber', () => {
	it('reads one stored Penjualan by its Nomor Struk', async () => {
		mock.onGet('/penjualan/7').reply(200, { data: { sale: { ...sale, receipt_number: 7 } } });

		await expect(penjualanApi.getByReceiptNumber(7)).resolves.toEqual({
			...sale,
			receipt_number: 7
		});
		expect(mock.history.get[0]!.url).toBe('/penjualan/7');
	});

	it('reads the items and the Pembayaran of the stored sale, Kembalian and all', async () => {
		mock.onGet('/penjualan/7').reply(200, { data: { sale: { ...sale, receipt_number: 7 } } });

		const found = await penjualanApi.getByReceiptNumber(7);

		expect(found.items).toEqual(sale.items);
		expect(found.payment).toEqual(sale.payment);
	});

	it('refuses a number that is not a Nomor Struk before the request is sent', async () => {
		mock.onGet('/penjualan/0').reply(200, { data: { sale } });

		await expect(penjualanApi.getByReceiptNumber(0)).rejects.toThrow();
		expect(mock.history.get).toHaveLength(0);
	});

	it('surfaces a Nomor Struk that names nothing as the readable 404 Go sent', async () => {
		mock.onGet('/penjualan/999').reply(404, {
			message: 'Penjualan tidak ditemukan.',
			error: 'sale_not_found'
		});

		await expect(penjualanApi.getByReceiptNumber(999)).rejects.toMatchObject({
			status: 404,
			code: 'sale_not_found',
			message: 'Penjualan tidak ditemukan.'
		});
	});

	it('fails loudly when the answer is not a Penjualan', async () => {
		mock.onGet('/penjualan/7').reply(200, { data: { sale: { receipt_number: 'tujuh' } } });

		await expect(penjualanApi.getByReceiptNumber(7)).rejects.toThrow();
	});
});

describe('penjualanApi.cetakStruk', () => {
	it('prints the Struk of one stored Penjualan and answers the outcome', async () => {
		mock.onPost('/penjualan/7/struk').reply(200, { data: { print: printed } });

		await expect(penjualanApi.cetakStruk(7)).resolves.toEqual(printed);
		expect(mock.history.post[0]!.url).toBe('/penjualan/7/struk');
		// A reprint carries nothing but the Nomor Struk in the path: the API reads
		// the sale and the template itself.
		expect(mock.history.post[0]!.data).toBeUndefined();
	});

	it('answers a print that failed as a result, not as an error', async () => {
		const gagal = { printed: false, message: 'Printer belum diatur.' };
		mock.onPost('/penjualan/7/struk').reply(200, { data: { print: gagal } });

		await expect(penjualanApi.cetakStruk(7)).resolves.toEqual(gagal);
	});

	it('refuses a number that is not a Nomor Struk before the request is sent', async () => {
		mock.onPost('/penjualan/0/struk').reply(200, { data: { print: printed } });

		await expect(penjualanApi.cetakStruk(0)).rejects.toThrow();
		expect(mock.history.post).toHaveLength(0);
	});

	it('surfaces a Nomor Struk that names nothing as the readable 404 Go sent', async () => {
		mock.onPost('/penjualan/999/struk').reply(404, {
			message: 'Penjualan tidak ditemukan.',
			error: 'sale_not_found'
		});

		await expect(penjualanApi.cetakStruk(999)).rejects.toMatchObject({
			status: 404,
			code: 'sale_not_found',
			message: 'Penjualan tidak ditemukan.'
		});
	});

	it('fails loudly when the answer is not a print result', async () => {
		mock.onPost('/penjualan/7/struk').reply(200, { data: { print: { printed: 'ya' } } });

		await expect(penjualanApi.cetakStruk(7)).rejects.toThrow();
	});
});

import MockAdapter from 'axios-mock-adapter';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { apiClient } from '$lib/api/client';
import { produkApi } from './produk.api';

let mock: MockAdapter;

const produk = {
	id: 1,
	name: 'Kopi Susu',
	code: 'KOPI-01',
	price: 18000,
	category: 'Minuman',
	stock: 12,
	active: true,
	sold: false
};

beforeEach(() => {
	mock = new MockAdapter(apiClient);
});

afterEach(() => {
	mock.restore();
});

describe('produkApi.list', () => {
	it('returns the catalogue and leaves filters it was not given off the URL', async () => {
		mock.onGet('/produk').reply(200, { data: [produk] });

		await expect(produkApi.list({ name: '', code: '', category: '' })).resolves.toEqual([produk]);
		expect(mock.history.get[0]!.params).toEqual({});
	});

	it('sends the filters it was given', async () => {
		mock.onGet('/produk').reply(200, { data: [] });

		await produkApi.list({ name: 'kopi', code: '', category: 'Minuman', active: true });

		expect(mock.history.get[0]!.params).toEqual({
			name: 'kopi',
			category: 'Minuman',
			active: 'true'
		});
	});

	it('sends active=false rather than dropping it as falsy', async () => {
		mock.onGet('/produk').reply(200, { data: [] });

		await produkApi.list({ name: '', code: '', category: '', active: false });

		expect(mock.history.get[0]!.params).toEqual({ active: 'false' });
	});

	it('surfaces a refused Peran as a normalized AppError', async () => {
		mock.onGet('/produk').reply(403, {
			message: 'Anda tidak berhak melakukan tindakan ini.',
			error: 'forbidden'
		});

		await expect(produkApi.list({ name: '', code: '', category: '' })).rejects.toMatchObject({
			status: 403,
			code: 'forbidden'
		});
	});

	it('fails loudly when the answer is not a catalogue', async () => {
		mock.onGet('/produk').reply(200, { data: { produk: 'bukan array' } });

		await expect(produkApi.list({ name: '', code: '', category: '' })).rejects.toThrow();
	});
});

describe('produkApi.categories', () => {
	it('returns the Kategori in use', async () => {
		mock.onGet('/produk/kategori').reply(200, { data: ['Makanan', 'Minuman'] });

		await expect(produkApi.categories()).resolves.toEqual(['Makanan', 'Minuman']);
	});
});

describe('produkApi.create', () => {
	it('posts a new Produk and returns the stored one', async () => {
		mock.onPost('/produk').reply(201, { data: { product: produk } });

		const created = await produkApi.create({
			name: '  Kopi Susu ',
			code: 'KOPI-01',
			price: 18000,
			category: 'Minuman',
			stock: 12
		});

		expect(created).toEqual(produk);
		expect(JSON.parse(mock.history.post[0]!.data as string)).toEqual({
			name: 'Kopi Susu',
			code: 'KOPI-01',
			price: 18000,
			category: 'Minuman',
			stock: 12
		});
	});

	it('sends null for a Kode and Kategori the form left blank', async () => {
		mock
			.onPost('/produk')
			.reply(201, { data: { product: { ...produk, code: null, category: null } } });

		await produkApi.create({ name: 'Air', code: '', price: 3000, category: '', stock: 0 });

		expect(JSON.parse(mock.history.post[0]!.data as string)).toEqual({
			name: 'Air',
			code: null,
			price: 3000,
			category: null,
			stock: 0
		});
	});

	it('posts the Status a new Produk was given', async () => {
		mock.onPost('/produk').reply(201, { data: { product: { ...produk, active: false } } });

		await produkApi.create({
			name: 'Belum Dijual',
			code: null,
			price: 1000,
			category: null,
			stock: 1,
			active: false
		});

		expect(JSON.parse(mock.history.post[0]!.data as string)).toEqual({
			name: 'Belum Dijual',
			code: null,
			price: 1000,
			category: null,
			stock: 1,
			active: false
		});
	});

	it('rejects an input the schema refuses before any request is made', async () => {
		await expect(
			produkApi.create({ name: 'Kopi', code: null, price: -1, category: null, stock: 0 })
		).rejects.toThrow();
		expect(mock.history.post).toHaveLength(0);
	});

	it('surfaces a taken Kode as a normalized AppError', async () => {
		mock.onPost('/produk').reply(409, {
			message: 'Kode sudah dipakai Produk lain.',
			error: 'code_taken'
		});

		await expect(
			produkApi.create({ name: 'Kopi', code: 'KOPI-01', price: 18000, category: null, stock: 0 })
		).rejects.toMatchObject({ status: 409, code: 'code_taken' });
	});
});

describe('produkApi.update', () => {
	it('puts the editable record to the Produk the path names, with no Stok in the body', async () => {
		mock.onPut('/produk/1').reply(200, { data: { product: { ...produk, price: 22000 } } });

		const updated = await produkApi.update(1, {
			name: 'Kopi Susu',
			code: 'KOPI-01',
			price: 22000,
			category: 'Minuman'
		});

		expect(updated.price).toBe(22000);

		// An edit cannot move Stok, so the request carries no field for one — not even
		// the value the Produk already has (ADR-0014).
		const body = JSON.parse(mock.history.put[0]!.data as string);
		expect(body).toEqual({
			name: 'Kopi Susu',
			code: 'KOPI-01',
			price: 22000,
			category: 'Minuman'
		});
		expect(body).not.toHaveProperty('stock');
	});

	it('drops a Stok and a Status that arrive from somewhere else', async () => {
		// The shape an older caller would hand it. The schema is the boundary, so the
		// request is what proves the fields never leave the app.
		mock.onPut('/produk/1').reply(200, { data: { product: produk } });

		await produkApi.update(1, {
			name: 'Kopi Susu',
			code: 'KOPI-01',
			price: 18000,
			category: 'Minuman',
			stock: 10,
			active: false
		} as never);

		expect(JSON.parse(mock.history.put[0]!.data as string)).toEqual({
			name: 'Kopi Susu',
			code: 'KOPI-01',
			price: 18000,
			category: 'Minuman'
		});
	});
});

describe('produkApi.setActive', () => {
	it('patches the Produk and returns its new state', async () => {
		mock.onPatch('/produk/1').reply(200, { data: { product: { ...produk, active: false } } });

		const updated = await produkApi.setActive(1, false);

		expect(updated.active).toBe(false);
		expect(JSON.parse(mock.history.patch[0]!.data as string)).toEqual({ active: false });
	});

	it('refuses a Status that is not a boolean instead of posting it', async () => {
		// The body crosses the HTTP boundary like every other one, so it is parsed
		// like every other one: `{ active: 'false' }` would reach Go as a string and
		// come back a 400 the UI could not explain.
		await expect(produkApi.setActive(1, 'false' as unknown as boolean)).rejects.toThrow();
		expect(mock.history.patch).toHaveLength(0);
	});
});

describe('produkApi.addStock', () => {
	it('posts the quantity to the Produk the path names and returns the stored one', async () => {
		mock.onPost('/produk/1/stok').reply(200, { data: { product: { ...produk, stock: 15 } } });

		const updated = await produkApi.addStock(1, 3);

		expect(updated.stock).toBe(15);
		// Only the quantity crosses the boundary, never a new total: sending the
		// total would let two restocks overwrite each other, since the Stok already
		// on the Produk is the API's to add to.
		expect(JSON.parse(mock.history.post[0]!.data as string)).toEqual({ quantity: 3 });
	});

	it('rejects a quantity of zero before any request is made', async () => {
		await expect(produkApi.addStock(1, 0)).rejects.toThrow();
		expect(mock.history.post).toHaveLength(0);
	});

	it('rejects a negative quantity before any request is made', async () => {
		await expect(produkApi.addStock(1, -3)).rejects.toThrow();
		expect(mock.history.post).toHaveLength(0);
	});

	it('surfaces a missing Produk as a normalized AppError', async () => {
		mock.onPost('/produk/404/stok').reply(404, {
			message: 'Produk tidak ditemukan.',
			error: 'product_not_found'
		});

		await expect(produkApi.addStock(404, 1)).rejects.toMatchObject({
			status: 404,
			code: 'product_not_found'
		});
	});

	it('fails loudly when the answer is not a Produk', async () => {
		mock
			.onPost('/produk/1/stok')
			.reply(200, { data: { product: { ...produk, stock: 'tiga belas' } } });

		await expect(produkApi.addStock(1, 3)).rejects.toThrow();
	});
});

describe('produkApi.lowStock', () => {
	it('returns the threshold with the Produk it selected', async () => {
		mock
			.onGet('/produk/stok-menipis')
			.reply(200, { data: { threshold: 5, products: [{ ...produk, stock: 2 }] } });

		const low = await produkApi.lowStock();

		expect(low.threshold).toBe(5);
		expect(low.products).toHaveLength(1);
		expect(low.products[0]!.stock).toBe(2);
	});

	it('reads a restocked store as an empty list', async () => {
		mock.onGet('/produk/stok-menipis').reply(200, { data: { threshold: 5, products: [] } });

		await expect(produkApi.lowStock()).resolves.toEqual({ threshold: 5, products: [] });
	});

	it('fails loudly when the answer carries no threshold', async () => {
		// Without the rule that selected the list the UI would have to invent one,
		// and the badge it showed would disagree with the list underneath it.
		mock.onGet('/produk/stok-menipis').reply(200, { data: { products: [] } });

		await expect(produkApi.lowStock()).rejects.toThrow();
	});

	it('surfaces a refused Peran as a normalized AppError', async () => {
		mock.onGet('/produk/stok-menipis').reply(403, {
			message: 'Anda tidak berhak melakukan tindakan ini.',
			error: 'forbidden'
		});

		await expect(produkApi.lowStock()).rejects.toMatchObject({ status: 403, code: 'forbidden' });
	});
});

describe('produkApi.remove', () => {
	it('asks the BFF to delete the Produk', async () => {
		mock.onDelete('/produk/1').reply(204);

		await expect(produkApi.remove(1)).resolves.toBeUndefined();
		expect(mock.history.delete).toHaveLength(1);
	});

	it('surfaces a refused delete as a normalized AppError', async () => {
		mock.onDelete('/produk/1').reply(409, {
			message: 'Produk yang sudah pernah terjual hanya bisa dinonaktifkan.',
			error: 'product_has_sales'
		});

		await expect(produkApi.remove(1)).rejects.toMatchObject({
			status: 409,
			code: 'product_has_sales'
		});
	});
});

import { describe, expect, it } from 'vitest';
import {
	KategoriListSchema,
	ProdukEnvelopeSchema,
	ProdukFilterSchema,
	ProdukInputSchema,
	ProdukListSchema,
	ProdukSchema
} from './produk.schema';

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

describe('ProdukSchema', () => {
	it('reads a Produk as the API answers it', () => {
		expect(ProdukSchema.parse(produk)).toEqual(produk);
	});

	it('reads an absent Kode and Kategori as null, not as a missing field', () => {
		const parsed = ProdukSchema.parse({ ...produk, code: null, category: null });

		expect(parsed.code).toBeNull();
		expect(parsed.category).toBeNull();
	});

	it('rejects a Produk that is missing its sales history', () => {
		// `sold` is what tells the UI whether a Produk may be deleted, so an answer
		// without it has to fail loudly rather than be read as "never sold".
		const withoutSold: Partial<typeof produk> = { ...produk };
		delete withoutSold.sold;

		expect(() => ProdukSchema.parse(withoutSold)).toThrow();
	});

	it('rejects a fractional Harga: money is whole rupiah in this app', () => {
		expect(() => ProdukSchema.parse({ ...produk, price: 18000.5 })).toThrow();
	});
});

describe('ProdukInputSchema', () => {
	it('trims text and coerces the numbers a form submits as strings', () => {
		const parsed = ProdukInputSchema.parse({
			name: '  Kopi Susu ',
			code: ' KOPI-01 ',
			price: '18000',
			category: ' Minuman ',
			stock: '12'
		});

		expect(parsed).toEqual({
			name: 'Kopi Susu',
			code: 'KOPI-01',
			price: 18000,
			category: 'Minuman',
			stock: 12
		});
	});

	it('turns blank Kode and Kategori into null, the value the API stores', () => {
		const parsed = ProdukInputSchema.parse({
			name: 'Air Mineral',
			code: '   ',
			price: 3000,
			category: '',
			stock: 0
		});

		expect(parsed.code).toBeNull();
		expect(parsed.category).toBeNull();
	});

	it('accepts a Kode and Kategori that are already null', () => {
		const parsed = ProdukInputSchema.parse({
			name: 'Air Mineral',
			code: null,
			price: 3000,
			category: null,
			stock: 0
		});

		expect(parsed.code).toBeNull();
		expect(parsed.category).toBeNull();
	});

	it('rejects a blank name with a message fit for the form', () => {
		const result = ProdukInputSchema.safeParse({ name: '   ', price: 1000, stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Nama Produk wajib diisi.');
	});

	it('rejects a negative Harga', () => {
		const result = ProdukInputSchema.safeParse({ name: 'Kopi', price: -1, stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga tidak boleh negatif.');
	});

	it('rejects a negative Stok', () => {
		const result = ProdukInputSchema.safeParse({ name: 'Kopi', price: 1000, stock: -1 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Stok tidak boleh negatif.');
	});

	it('rejects a Harga that is not a number instead of reading it as 0', () => {
		const result = ProdukInputSchema.safeParse({ name: 'Kopi', price: 'seribu', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('rejects a blank Harga instead of quietly pricing the Produk at 0', () => {
		const result = ProdukInputSchema.safeParse({ name: 'Kopi', price: '', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('rejects a fractional Harga', () => {
		const result = ProdukInputSchema.safeParse({ name: 'Kopi', price: '18000.5', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('rejects a Harga written with the Indonesian thousands separator instead of reading it as 18', () => {
		// `Number('18.000')` is 18, so coercing would have saved this Produk at Rp 18
		// while the list showed "Rp 18.000" back — a silent 1000× mistake.
		const result = ProdukInputSchema.safeParse({ name: 'Kopi', price: '18.000', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('rejects a separator wherever the grouping lands', () => {
		// "1.500" used to coerce to 1.5 and "18,000" to NaN, so the same typo failed
		// or not depending on the digits. Money has no decimals here, so all of them
		// have to fail the same way.
		for (const price of ['1.500', '18,000', '1 500', '1e3']) {
			expect(ProdukInputSchema.safeParse({ name: 'Kopi', price, stock: 0 }).success).toBe(false);
		}
	});

	it('rejects a Stok with a separator, since it is the same rule', () => {
		const result = ProdukInputSchema.safeParse({ name: 'Kopi', price: 1000, stock: '1.000' });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Stok harus bilangan bulat.');
	});

	it('rejects an absent Harga instead of coercing it to 0', () => {
		const result = ProdukInputSchema.safeParse({ name: 'Kopi', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('accepts whitespace around a whole Harga', () => {
		expect(ProdukInputSchema.parse({ name: 'Kopi', price: ' 18000 ', stock: 0 }).price).toBe(18000);
	});

	it('accepts a Harga of 0: a free Produk is a choice, not a mistake', () => {
		expect(ProdukInputSchema.parse({ name: 'Air', price: '0', stock: '0' }).price).toBe(0);
	});
});

describe('ProdukFilterSchema', () => {
	it('defaults every text filter to empty and leaves active unfiltered', () => {
		expect(ProdukFilterSchema.parse({})).toEqual({ name: '', code: '', category: '' });
	});

	it('trims the filters it is given', () => {
		const parsed = ProdukFilterSchema.parse({ name: ' kopi ', category: ' Minuman ' });

		expect(parsed.name).toBe('kopi');
		expect(parsed.category).toBe('Minuman');
	});

	it('keeps an explicit active filter, including false', () => {
		expect(ProdukFilterSchema.parse({ active: false }).active).toBe(false);
		expect(ProdukFilterSchema.parse({ active: true }).active).toBe(true);
	});
});

describe('ProdukEnvelopeSchema', () => {
	it('reads the Produk the BFF answers with', () => {
		expect(ProdukEnvelopeSchema.parse({ data: { product: produk } }).data.product).toEqual(produk);
	});

	it('rejects an answer that carries no Produk', () => {
		expect(() => ProdukEnvelopeSchema.parse({ data: {} })).toThrow();
	});
});

describe('ProdukListSchema', () => {
	it('reads a catalogue', () => {
		expect(ProdukListSchema.parse({ data: [produk] }).data).toHaveLength(1);
	});

	it('reads an empty catalogue as an empty list, not a missing field', () => {
		expect(ProdukListSchema.parse({ data: [] }).data).toEqual([]);
	});
});

describe('KategoriListSchema', () => {
	it('reads the Kategori in use', () => {
		expect(KategoriListSchema.parse({ data: ['Makanan', 'Minuman'] }).data).toEqual([
			'Makanan',
			'Minuman'
		]);
	});

	it('reads a store with no Kategori yet as an empty list', () => {
		expect(KategoriListSchema.parse({ data: [] }).data).toEqual([]);
	});
});

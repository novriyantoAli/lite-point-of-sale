import { describe, expect, it } from 'vitest';
import {
	KategoriListSchema,
	ProdukEnvelopeSchema,
	ProdukFilterSchema,
	CreateProdukInputSchema,
	UpdateProdukInputSchema,
	ProdukListSchema,
	ProdukSchema,
	SetActiveInputSchema,
	StokMenipisSchema,
	TambahStokInputSchema
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

describe('CreateProdukInputSchema', () => {
	it('trims text and coerces the numbers a form submits as strings', () => {
		const parsed = CreateProdukInputSchema.parse({
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
		const parsed = CreateProdukInputSchema.parse({
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
		const parsed = CreateProdukInputSchema.parse({
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
		const result = CreateProdukInputSchema.safeParse({ name: '   ', price: 1000, stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Nama Produk wajib diisi.');
	});

	it('rejects a negative Harga', () => {
		const result = CreateProdukInputSchema.safeParse({ name: 'Kopi', price: -1, stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga tidak boleh negatif.');
	});

	it('rejects a negative Stok', () => {
		const result = CreateProdukInputSchema.safeParse({ name: 'Kopi', price: 1000, stock: -1 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Stok tidak boleh negatif.');
	});

	it('rejects a Harga that is not a number instead of reading it as 0', () => {
		const result = CreateProdukInputSchema.safeParse({ name: 'Kopi', price: 'seribu', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('rejects a blank Harga instead of quietly pricing the Produk at 0', () => {
		const result = CreateProdukInputSchema.safeParse({ name: 'Kopi', price: '', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('rejects a fractional Harga', () => {
		const result = CreateProdukInputSchema.safeParse({ name: 'Kopi', price: '18000.5', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('rejects a Harga written with the Indonesian thousands separator instead of reading it as 18', () => {
		// `Number('18.000')` is 18, so coercing would have saved this Produk at Rp 18
		// while the list showed "Rp 18.000" back — a silent 1000× mistake.
		const result = CreateProdukInputSchema.safeParse({ name: 'Kopi', price: '18.000', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('rejects a separator wherever the grouping lands', () => {
		// "1.500" used to coerce to 1.5 and "18,000" to NaN, so the same typo failed
		// or not depending on the digits. Money has no decimals here, so all of them
		// have to fail the same way.
		for (const price of ['1.500', '18,000', '1 500', '1e3']) {
			expect(CreateProdukInputSchema.safeParse({ name: 'Kopi', price, stock: 0 }).success).toBe(
				false
			);
		}
	});

	it('rejects a Stok with a separator, since it is the same rule', () => {
		const result = CreateProdukInputSchema.safeParse({ name: 'Kopi', price: 1000, stock: '1.000' });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Stok harus bilangan bulat.');
	});

	it('rejects an absent Harga instead of coercing it to 0', () => {
		const result = CreateProdukInputSchema.safeParse({ name: 'Kopi', stock: 0 });

		expect(result.success).toBe(false);
		expect(result.error?.issues[0]?.message).toBe('Harga harus bilangan bulat.');
	});

	it('accepts whitespace around a whole Harga', () => {
		expect(CreateProdukInputSchema.parse({ name: 'Kopi', price: ' 18000 ', stock: 0 }).price).toBe(
			18000
		);
	});

	it('accepts a Harga of 0: a free Produk is a choice, not a mistake', () => {
		expect(CreateProdukInputSchema.parse({ name: 'Air', price: '0', stock: '0' }).price).toBe(0);
	});

	it('carries the Status a new Produk was given', () => {
		expect(
			CreateProdukInputSchema.parse({ name: 'Kopi', price: 1000, stock: 0, active: false }).active
		).toBe(false);
	});

	it('leaves the Status out when it was not asked for, so a create defaults to Aktif', () => {
		// An absent key is what lets the Go side own the default: a Produk added with
		// no Status asked for is Aktif.
		const parsed = CreateProdukInputSchema.parse({ name: 'Kopi', price: 1000, stock: 0 });

		expect('active' in parsed).toBe(false);
		expect(JSON.stringify(parsed)).not.toContain('active');
	});

	it('rejects a Status that arrived as a string', () => {
		expect(() =>
			CreateProdukInputSchema.parse({ name: 'Kopi', price: 1000, stock: 0, active: 'true' })
		).toThrow();
	});
});

describe('UpdateProdukInputSchema', () => {
	it('carries the editable record, and nothing else', () => {
		const parsed = UpdateProdukInputSchema.parse({
			name: '  Kopi Susu ',
			code: ' KOPI-01 ',
			price: '22000',
			category: ' Minuman '
		});

		expect(parsed).toEqual({
			name: 'Kopi Susu',
			code: 'KOPI-01',
			price: 22000,
			category: 'Minuman'
		});
	});

	it('cannot carry a Stok into an edit, even when a form hands it one', () => {
		// The shape a form would send if it still held a Stok field: the number it read
		// when it opened. The schema has nowhere to put it, so it never reaches the
		// request — which is what stops an edit from overwriting a later delivery
		// (ADR-0014).
		const parsed = UpdateProdukInputSchema.parse({ name: 'Kopi Susu', price: 22000, stock: 10 });

		expect('stock' in parsed).toBe(false);
		expect(JSON.stringify(parsed)).not.toContain('stock');
	});

	it('cannot carry a Status into an edit either', () => {
		const parsed = UpdateProdukInputSchema.parse({
			name: 'Kopi Susu',
			price: 22000,
			active: false
		});

		expect('active' in parsed).toBe(false);
	});

	it('shares the Nama and Harga rules with a create, so the two cannot disagree', () => {
		expect(
			UpdateProdukInputSchema.safeParse({ name: '   ', price: 1000 }).error?.issues[0]?.message
		).toBe('Nama Produk wajib diisi.');
		expect(
			UpdateProdukInputSchema.safeParse({ name: 'Kopi', price: -1 }).error?.issues[0]?.message
		).toBe('Harga tidak boleh negatif.');
		expect(UpdateProdukInputSchema.parse({ name: ' Kopi ', price: ' 22000 ' })).toEqual({
			name: 'Kopi',
			code: null,
			price: 22000,
			category: null
		});
	});
});

describe('SetActiveInputSchema', () => {
	it('takes the Status the button posts', () => {
		expect(SetActiveInputSchema.parse({ active: false })).toEqual({ active: false });
		expect(SetActiveInputSchema.parse({ active: true })).toEqual({ active: true });
	});

	it('rejects a body that forgot the Status', () => {
		expect(() => SetActiveInputSchema.parse({})).toThrow();
	});

	it('rejects a Status that arrived as a string', () => {
		// A checkbox or a query string can hand over "false", which is truthy to
		// anything that only asks whether it is empty.
		expect(() => SetActiveInputSchema.parse({ active: 'false' })).toThrow();
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

describe('TambahStokInputSchema', () => {
	it('reads the quantity a form submits as text', () => {
		expect(TambahStokInputSchema.parse({ quantity: '5' })).toEqual({ quantity: 5 });
		expect(TambahStokInputSchema.parse({ quantity: 5 })).toEqual({ quantity: 5 });
	});

	it('accepts whitespace around a whole quantity', () => {
		expect(TambahStokInputSchema.parse({ quantity: ' 12 ' })).toEqual({ quantity: 12 });
	});

	it('rejects a quantity of zero with a message fit for the form', () => {
		// A restock of nothing is not a restock: it would record an event that did
		// not happen, so the form asks again instead of accepting it.
		expect(() => TambahStokInputSchema.parse({ quantity: '0' })).toThrow(
			'Jumlah Stok harus lebih dari nol.'
		);
	});

	it('rejects a negative quantity: Stok leaves through a Penjualan, not a form', () => {
		expect(() => TambahStokInputSchema.parse({ quantity: '-3' })).toThrow(
			'Jumlah Stok harus lebih dari nol.'
		);
	});

	it('rejects a blank quantity instead of reading it as 0', () => {
		expect(() => TambahStokInputSchema.parse({ quantity: '' })).toThrow();
	});

	it('rejects an absent quantity', () => {
		expect(() => TambahStokInputSchema.parse({})).toThrow();
	});

	it('rejects a fractional quantity: Stok is whole units', () => {
		expect(() => TambahStokInputSchema.parse({ quantity: '2.5' })).toThrow(
			'Jumlah Stok harus bilangan bulat.'
		);
	});

	it('rejects the Indonesian thousands separator, the same rule as Harga', () => {
		// `Number('1.000')` is 1, so coercing would restock a thousandth of what
		// the Admin typed and show the mistake nowhere.
		expect(() => TambahStokInputSchema.parse({ quantity: '1.000' })).toThrow(
			'Jumlah Stok harus bilangan bulat.'
		);
	});
});

describe('StokMenipisSchema', () => {
	it('reads the threshold together with the Produk it selected', () => {
		const parsed = StokMenipisSchema.parse({ data: { threshold: 5, products: [produk] } });

		expect(parsed.data.threshold).toBe(5);
		expect(parsed.data.products).toEqual([produk]);
	});

	it('reads a restocked store as an empty list', () => {
		expect(StokMenipisSchema.parse({ data: { threshold: 5, products: [] } }).data.products).toEqual(
			[]
		);
	});

	it('rejects an answer without the threshold it selected by', () => {
		// The rule that decided the list is part of the answer: without it the UI
		// would have to invent one, and the badge it showed would be a second rule.
		expect(() => StokMenipisSchema.parse({ data: { products: [] } })).toThrow();
	});
});

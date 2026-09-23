import { beforeEach, describe, expect, it } from 'vitest';
import type { Produk } from '$lib/domains/produk';
import { keranjangState } from './keranjang.state.svelte';

/** A catalogue row, with only what a test cares about overridden. */
function produk(overrides: Partial<Produk> = {}): Produk {
	return {
		id: 1,
		name: 'Kopi Susu',
		code: 'KOPI-01',
		price: 18000,
		category: null,
		stock: 10,
		active: true,
		sold: false,
		...overrides
	};
}

const teh = produk({ id: 2, name: 'Teh Botol', code: null, price: 6000, stock: 4 });

beforeEach(() => {
	keranjangState.clear();
});

describe('KeranjangState', () => {
	it('starts empty, with a total of nothing', () => {
		expect(keranjangState.items).toEqual([]);
		expect(keranjangState.units).toBe(0);
		expect(keranjangState.total).toBe(0);
		expect(keranjangState.exceedsStock).toBe(false);
	});

	it('adds one unit of a Produk', () => {
		keranjangState.add(produk());

		expect(keranjangState.items).toHaveLength(1);
		expect(keranjangState.items[0]!.qty).toBe(1);
		expect(keranjangState.units).toBe(1);
		expect(keranjangState.total).toBe(18000);
	});

	it('raises the quantity of a Produk that is already in the keranjang', () => {
		keranjangState.add(produk());
		keranjangState.add(produk());

		// One Produk is one line (CONTEXT.md, Item) — never two lines of the same
		// Produk, which the API would have to merge anyway.
		expect(keranjangState.items).toHaveLength(1);
		expect(keranjangState.items[0]!.qty).toBe(2);
		expect(keranjangState.total).toBe(36000);
	});

	it('adds several units at once', () => {
		keranjangState.add(produk(), 3);

		expect(keranjangState.items[0]!.qty).toBe(3);
		expect(keranjangState.units).toBe(3);
	});

	it('keeps one line per Produk and sums them in the total', () => {
		keranjangState.add(produk());
		keranjangState.add(teh, 3);

		expect(keranjangState.items).toHaveLength(2);
		expect(keranjangState.units).toBe(4);
		expect(keranjangState.total).toBe(18000 + 3 * 6000);
	});

	it('takes the freshest Harga and Stok for a Produk it already holds', () => {
		keranjangState.add(produk());
		keranjangState.add(produk({ price: 20000, stock: 3 }));

		expect(keranjangState.items[0]!.produk.price).toBe(20000);
		expect(keranjangState.items[0]!.produk.stock).toBe(3);
		// The quantity is still two — the second add raised it — but both units are
		// now priced at what the catalogue last reported.
		expect(keranjangState.total).toBe(2 * 20000);
	});

	it('sets a line to a whole quantity of at least one', () => {
		keranjangState.add(produk());

		expect(keranjangState.setQuantity(1, 5)).toBe(5);
		expect(keranjangState.items[0]!.qty).toBe(5);
		expect(keranjangState.total).toBe(5 * 18000);
	});

	it('ignores a quantity that is not a whole number of at least one, and says what it kept', () => {
		keranjangState.add(produk());

		for (const rejected of [0, -2, 1.5, Number.NaN, Number.POSITIVE_INFINITY]) {
			expect(keranjangState.setQuantity(1, rejected)).toBe(1);
		}
		expect(keranjangState.items[0]!.qty).toBe(1);
	});

	it('accepts a quantity above the Stok, so the line can be seen and fixed', () => {
		keranjangState.add(produk({ stock: 2 }), 3);

		expect(keranjangState.items[0]!.qty).toBe(3);
		expect(keranjangState.exceedsStock).toBe(true);
	});

	it('reports no line above the Stok when every line fits', () => {
		keranjangState.add(produk({ stock: 2 }), 2);
		keranjangState.add(teh, 4);

		expect(keranjangState.exceedsStock).toBe(false);
	});

	it('removes one line and leaves the rest alone', () => {
		keranjangState.add(produk());
		keranjangState.add(teh);

		keranjangState.remove(1);

		expect(keranjangState.items).toHaveLength(1);
		expect(keranjangState.items[0]!.produk.id).toBe(2);
		expect(keranjangState.total).toBe(6000);
	});

	it('empties the keranjang', () => {
		keranjangState.add(produk(), 2);
		keranjangState.add(teh);

		keranjangState.clear();

		expect(keranjangState.items).toEqual([]);
		expect(keranjangState.total).toBe(0);
	});
});

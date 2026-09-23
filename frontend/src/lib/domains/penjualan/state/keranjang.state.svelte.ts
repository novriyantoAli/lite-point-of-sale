import type { Produk } from '$lib/domains/produk';

/**
 * One line of the keranjang: a Produk and how many units of it are being sold.
 * The Produk is kept whole so the line can show its name, Kode, Harga and
 * remaining Stok without a second lookup.
 */
export interface KeranjangItem {
	produk: Produk;
	qty: number;
}

/**
 * The keranjang, as UI-only state (ADR-0006): a draft selection of Produk and
 * quantities is not server data, so it lives in runes and never in the query
 * cache. The Stok and Harga it shows did come from the catalogue, but the API
 * re-reads both when the sale is written — so a stale line can make the total on
 * screen wrong for a moment, never make the Penjualan wrong.
 *
 * A module singleton rather than component state, so a Kasir who steps over to
 * the catalogue and comes back finds the keranjang they were building still
 * there. It is cleared deliberately, by checkout or by the Kosongkan button.
 */
class KeranjangState {
	items = $state<KeranjangItem[]>([]);

	/** The running total the Kasir reads out to the buyer, in whole rupiah. */
	get total(): number {
		return this.items.reduce((sum, item) => sum + item.produk.price * item.qty, 0);
	}

	/** How many units the keranjang holds, across every line. */
	get units(): number {
		return this.items.reduce((sum, item) => sum + item.qty, 0);
	}

	/**
	 * Whether any line asks for more than the Stok the catalogue reported. The
	 * till blocks checkout on this before the request is sent, and the API refuses
	 * it too — the Stok can have moved since this screen read it.
	 */
	get exceedsStock(): boolean {
		return this.items.some((item) => item.qty > item.produk.stock);
	}

	/**
	 * Adds units of a Produk. One Produk is one line (CONTEXT.md, Item), so adding
	 * a Produk that is already in the keranjang raises that line's quantity rather
	 * than appending a second line for the same Produk.
	 *
	 * The catalogue row passed in replaces the one already stored, so a line
	 * carries the freshest Harga and Stok this screen has seen.
	 */
	add(produk: Produk, qty = 1) {
		const existing = this.items.find((item) => item.produk.id === produk.id);
		if (existing) {
			existing.produk = produk;
			existing.qty += qty;
			return;
		}

		this.items.push({ produk, qty });
	}

	/**
	 * Sets a line's quantity and answers the quantity that is now stored. A
	 * quantity that is not a whole number of at least one is ignored rather than
	 * stored: a line of zero is not an Item, and answering what is stored lets the
	 * field write back the number the keranjang actually holds.
	 *
	 * A quantity above the Stok is accepted — the Kasir has to be able to see the
	 * line that is too big in order to fix it, and `exceedsStock` is what stops it
	 * from being checked out.
	 */
	setQuantity(productId: number, qty: number): number {
		const item = this.items.find((line) => line.produk.id === productId);
		if (!item) {
			return 0;
		}

		if (Number.isInteger(qty) && qty >= 1) {
			item.qty = qty;
		}

		return item.qty;
	}

	/** Removes one line. */
	remove(productId: number) {
		this.items = this.items.filter((item) => item.produk.id !== productId);
	}

	/** Empties the keranjang, for the Kosongkan button and after a checkout. */
	clear() {
		this.items = [];
	}
}

export const keranjangState = new KeranjangState();

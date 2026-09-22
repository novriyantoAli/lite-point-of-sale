import type { ProdukFilter } from '../schemas/produk.schema';

/**
 * The catalogue filters, as UI-only state (ADR-0006): what the Admin typed into
 * the filter bar is not server data, so it lives in runes and never in the query
 * cache. The query *reads* this state; it never owns it.
 *
 * A module singleton rather than component state, so the filters survive
 * navigating away from the catalogue and back — an Admin who filtered to one
 * Kategori and looked up a Produk comes back to the list they were reading.
 */
class ProdukFilterState {
	name = $state('');
	code = $state('');
	category = $state('');
	/**
	 * `null` is "Semua": the filter that is not set, which is neither `true` nor
	 * `false`. Go reads an absent `active` as "every Produk", Nonaktif included,
	 * so the two states must stay distinct all the way to the query string.
	 */
	active = $state<boolean | null>(null);

	/** The shape `produkApi.list` takes, with "Semua" as the absent filter. */
	get filter(): ProdukFilter {
		return {
			name: this.name,
			code: this.code,
			category: this.category,
			active: this.active ?? undefined
		};
	}

	/** Back to "Semua", for the filter bar's reset and for tests to start clean. */
	reset() {
		this.name = '';
		this.code = '';
		this.category = '';
		this.active = null;
	}
}

export const produkFilterState = new ProdukFilterState();

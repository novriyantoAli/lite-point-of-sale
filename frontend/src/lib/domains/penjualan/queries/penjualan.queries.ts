import { browser } from '$app/environment';
import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import { produkKeys } from '$lib/domains/produk';
import { penjualanApi } from '../api/penjualan.api';
import type { CheckoutInput, NomorStruk, Penjualan } from '../schemas/penjualan.schema';

/**
 * The cache keys of the Penjualan domain, in the one scheme every domain shares:
 * `["<domain>", "<shape>", ...]`. The mutation invalidates the subtree, and the
 * lookup screen (#29) reads `detail`; the sales list (#9) builds on the same
 * root rather than inventing a second scheme.
 */
export const penjualanKeys = {
	all: ['penjualan'] as const,
	detail: (nomorStruk: NomorStruk) => [...penjualanKeys.all, 'detail', nomorStruk] as const
};

/**
 * One stored Penjualan, read by its Nomor Struk. `nomorStruk` is a thunk so the
 * query re-derives when the lookup screen's field changes (§6.3), and the
 * factory is only mounted once a search has been submitted — there is no
 * "disabled" key for a number nobody asked for.
 *
 * `enabled: browser` for the same reason as the other domains: it is a
 * same-origin call to the BFF, so it can only be made where the BFF is
 * reachable, and a lookup does not need SSR.
 */
export function createPenjualanDetailQuery(nomorStruk: () => NomorStruk) {
	return createQuery<Penjualan, AppError>(() => ({
		queryKey: penjualanKeys.detail(nomorStruk()),
		queryFn: () => penjualanApi.getByReceiptNumber(nomorStruk()),
		enabled: browser
	}));
}

/**
 * Checks a cart out. On success it invalidates the whole `produk` subtree, not
 * just this domain's: the sale took Stok out and marked those Produk as sold, so
 * the catalogue the till was reading is stale — and so is the Stok-menipis list,
 * which a sale can add a Produk to.
 *
 * The sale itself is *not* written into the cache. The till shows the Penjualan
 * the mutation returned, and the lookup screen reads it again by its Nomor Struk
 * when it is asked for; seeding a `detail` entry here would be a second copy of
 * a record the reader fetches on its own.
 */
export function createCheckoutMutation() {
	const queryClient = useQueryClient();

	return createMutation<Penjualan, AppError, CheckoutInput>(() => ({
		mutationFn: (input: CheckoutInput) => penjualanApi.checkout(input),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: produkKeys.all });
			queryClient.invalidateQueries({ queryKey: penjualanKeys.all });
		}
	}));
}

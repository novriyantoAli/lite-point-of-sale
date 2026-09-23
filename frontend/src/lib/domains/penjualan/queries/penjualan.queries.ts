import { createMutation, useQueryClient } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import { produkKeys } from '$lib/domains/produk';
import { penjualanApi } from '../api/penjualan.api';
import type { CheckoutInput, Penjualan } from '../schemas/penjualan.schema';

/**
 * The cache keys of the Penjualan domain, in the one scheme every domain shares:
 * `["<domain>", "<shape>", ...]`. There is no reader of a Penjualan yet — the
 * till shows the sale the checkout mutation answers with — so the keys are here
 * for the mutation to invalidate and for the reprint (#8) and sales list (#9) to
 * build on.
 */
export const penjualanKeys = {
	all: ['penjualan'] as const
};

/**
 * Checks a cart out. On success it invalidates the whole `produk` subtree, not
 * just this domain's: the sale took Stok out and marked those Produk as sold, so
 * the catalogue the till was reading is stale — and so is the Stok-menipis list,
 * which a sale can add a Produk to.
 *
 * The sale itself is *not* written into the cache. The till shows the Penjualan
 * the mutation returned, and a reader that wants it again asks for it by its
 * Nomor Struk; seeding a cache entry here would be a second copy of a record
 * nothing has read yet.
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

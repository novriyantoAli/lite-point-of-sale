import { browser } from '$app/environment';
import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import { produkApi } from '../api/produk.api';
import type {
	CreateProdukInput,
	Produk,
	ProdukFilter,
	StokMenipis,
	UpdateProdukInput
} from '../schemas/produk.schema';

/**
 * The cache keys of the Produk domain, in the one scheme every domain shares:
 * `["<domain>", "<shape>", ...]`. Keeping them here rather than at the call site
 * is what makes invalidation predictable — a mutation invalidates a key, never a
 * string it typed by hand (ADR-0006).
 */
export const produkKeys = {
	all: ['produk'] as const,
	list: (filter: ProdukFilter) => [...produkKeys.all, 'list', filter] as const,
	kategori: () => [...produkKeys.all, 'kategori'] as const,
	stokMenipis: () => [...produkKeys.all, 'stok-menipis'] as const
};

/**
 * The catalogue, narrowed by the filters the Admin set. `filter` is a thunk so
 * the query re-derives whenever the runes state behind it changes — passing a
 * plain object would freeze it at the first render (§6.3).
 *
 * `enabled: browser` for the same reason as the Pengguna list: this is a
 * same-origin call to the BFF, so it can only be made where the BFF is
 * reachable, and the catalogue does not need SSR.
 */
export function createProdukListQuery(filter: () => ProdukFilter) {
	return createQuery<Produk[], AppError>(() => ({
		queryKey: produkKeys.list(filter()),
		queryFn: () => produkApi.list(filter()),
		enabled: browser
	}));
}

/**
 * The Kategori already in use, so the filter offers the ones that exist instead
 * of a list the frontend would have to maintain by hand.
 */
export function createKategoriListQuery() {
	return createQuery<string[], AppError>(() => ({
		queryKey: produkKeys.kategori(),
		queryFn: () => produkApi.categories(),
		enabled: browser
	}));
}

/**
 * The Produk to restock, with the threshold that selected them. It is a query of
 * its own rather than a filter on the list because the rule that decides it is
 * the domain's (`LowStockThreshold` in Go), not a filter the Admin sets — and
 * because the answer carries that rule back, so the screen can show it.
 */
export function createStokMenipisQuery() {
	return createQuery<StokMenipis, AppError>(() => ({
		queryKey: produkKeys.stokMenipis(),
		queryFn: () => produkApi.lowStock(),
		enabled: browser
	}));
}

/**
 * Every mutation invalidates `produkKeys.all`, not just the list: adding or
 * changing a Produk can introduce a Kategori the filter dropdown has not seen
 * yet, and deleting one can remove the last Produk that used it. Invalidating
 * the subtree keeps the two answers consistent instead of letting the dropdown
 * drift from the catalogue it describes.
 */

export function createProdukMutation() {
	const queryClient = useQueryClient();

	return createMutation<Produk, AppError, CreateProdukInput>(() => ({
		mutationFn: (input: CreateProdukInput) => produkApi.create(input),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: produkKeys.all })
	}));
}

/**
 * Changes the editable record of one Produk. Its input has no Stok — the domain's
 * `UpdateProdukInput` is where that is decided, and this signature follows it
 * (ADR-0014).
 */
export function createUpdateProdukMutation() {
	const queryClient = useQueryClient();

	return createMutation<Produk, AppError, { id: number; input: UpdateProdukInput }>(() => ({
		mutationFn: ({ id, input }) => produkApi.update(id, input),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: produkKeys.all })
	}));
}

export function createSetProdukActiveMutation() {
	const queryClient = useQueryClient();

	return createMutation<Produk, AppError, { id: number; active: boolean }>(() => ({
		mutationFn: ({ id, active }) => produkApi.setActive(id, active),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: produkKeys.all })
	}));
}

export function createDeleteProdukMutation() {
	const queryClient = useQueryClient();

	return createMutation<void, AppError, number>(() => ({
		mutationFn: (id: number) => produkApi.remove(id),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: produkKeys.all })
	}));
}

/**
 * Records a restock. It invalidates the whole `produk` subtree for the same
 * reason the other mutations do — and one more: a restock changes both the
 * catalogue's Stok column and the restock list, and the Produk it lifts above
 * the threshold has to leave that list. Invalidating only one of the two would
 * leave the screen arguing with itself.
 */
export function createAddStokMutation() {
	const queryClient = useQueryClient();

	return createMutation<Produk, AppError, { id: number; quantity: number }>(() => ({
		mutationFn: ({ id, quantity }) => produkApi.addStock(id, quantity),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: produkKeys.all })
	}));
}

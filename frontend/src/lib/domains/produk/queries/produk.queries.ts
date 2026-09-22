import { browser } from '$app/environment';
import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import { produkApi } from '../api/produk.api';
import type { Produk, ProdukFilter, ProdukInput } from '../schemas/produk.schema';

/**
 * The cache keys of the Produk domain, in the one scheme every domain shares:
 * `["<domain>", "<shape>", ...]`. Keeping them here rather than at the call site
 * is what makes invalidation predictable — a mutation invalidates a key, never a
 * string it typed by hand (ADR-0006).
 */
export const produkKeys = {
	all: ['produk'] as const,
	list: (filter: ProdukFilter) => [...produkKeys.all, 'list', filter] as const,
	kategori: () => [...produkKeys.all, 'kategori'] as const
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
 * Every mutation invalidates `produkKeys.all`, not just the list: adding or
 * changing a Produk can introduce a Kategori the filter dropdown has not seen
 * yet, and deleting one can remove the last Produk that used it. Invalidating
 * the subtree keeps the two answers consistent instead of letting the dropdown
 * drift from the catalogue it describes.
 */

export function createProdukMutation() {
	const queryClient = useQueryClient();

	return createMutation<Produk, AppError, ProdukInput>(() => ({
		mutationFn: (input: ProdukInput) => produkApi.create(input),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: produkKeys.all })
	}));
}

export function createUpdateProdukMutation() {
	const queryClient = useQueryClient();

	return createMutation<Produk, AppError, { id: number; input: ProdukInput }>(() => ({
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

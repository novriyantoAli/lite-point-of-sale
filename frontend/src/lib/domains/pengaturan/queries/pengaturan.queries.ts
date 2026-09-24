import { browser } from '$app/environment';
import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import { produkKeys } from '$lib/domains/produk';
import { pengaturanApi } from '../api/pengaturan.api';
import type { Pengaturan, UpdatePengaturanInput } from '../schemas/pengaturan.schema';

/**
 * The cache keys of the pengaturan domain, in the one scheme every domain
 * shares: `["<domain>", "<shape>", ...]`.
 */
export const pengaturanKeys = {
	all: ['pengaturan'] as const,
	current: () => [...pengaturanKeys.all, 'current'] as const
};

/**
 * The store's single row of Pengaturan. `enabled: browser` for the same reason
 * as the other domains: it is a same-origin call to the BFF, so it can only be
 * made where the BFF is reachable, and the form does not need SSR.
 */
export function createPengaturanQuery() {
	return createQuery<Pengaturan, AppError>(() => ({
		queryKey: pengaturanKeys.current(),
		queryFn: () => pengaturanApi.get(),
		enabled: browser
	}));
}

/**
 * Changes the Pengaturan. It invalidates its own subtree, and also the produk
 * domain's Stok-menipis report: the ambang Stok menipis lives here but the list
 * it selects is the produk domain's, so a change has to drop that cached list —
 * otherwise the /stok screen would keep showing the old ambang until it
 * naturally went stale (ADR-0017, keputusan 3). `produkKeys` is imported
 * through the produk barrel, which is the only cross-domain surface (ADR-0006).
 */
export function createUpdatePengaturanMutation() {
	const queryClient = useQueryClient();

	return createMutation<Pengaturan, AppError, UpdatePengaturanInput>(() => ({
		mutationFn: (input: UpdatePengaturanInput) => pengaturanApi.update(input),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: pengaturanKeys.all });
			queryClient.invalidateQueries({ queryKey: produkKeys.stokMenipis() });
		}
	}));
}

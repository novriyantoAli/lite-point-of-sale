import { browser } from '$app/environment';
import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import { produkKeys } from '$lib/domains/produk';
import { penjualanApi } from '../api/penjualan.api';
import type {
	CheckoutInput,
	HasilCetak,
	HasilCheckout,
	NomorStruk,
	OmzetHarian,
	Penjualan,
	PenjualanRingkas,
	TanggalLaporan
} from '../schemas/penjualan.schema';

/**
 * The cache keys of the Penjualan domain, in the one scheme every domain shares:
 * `["<domain>", "<shape>", ...]`. The mutation invalidates the subtree, and the
 * lookup screen (#29) reads `detail`; the sales list and the omzet report (#9)
 * build on the same root rather than inventing a second scheme.
 */
export const penjualanKeys = {
	all: ['penjualan'] as const,
	detail: (nomorStruk: NomorStruk) => [...penjualanKeys.all, 'detail', nomorStruk] as const,
	list: (tanggal: TanggalLaporan) => [...penjualanKeys.all, 'list', tanggal] as const,
	omzet: (tanggal: TanggalLaporan) => [...penjualanKeys.all, 'omzet', tanggal] as const
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
		enabled: browser,
		// A 404 is an answer, not a hiccup: that Nomor Struk names no Penjualan, and
		// asking Go again only delays the message the screen already has. Every
		// other failure (502, a dropped connection) keeps the app's one retry —
		// the `retry: 1` the root QueryClient sets (ADR-0006 conventions).
		retry: (failureCount, error) => error.status !== 404 && failureCount < 1
	}));
}

/**
 * The Penjualan of one store-local day, newest Nomor Struk first: the sales list
 * of the Laporan screen (#9). `tanggal` is a thunk so the query re-derives when
 * the day the Admin picked changes (§6.3).
 */
export function createPenjualanListQuery(tanggal: () => TanggalLaporan) {
	return createQuery<PenjualanRingkas[], AppError>(() => ({
		queryKey: penjualanKeys.list(tanggal()),
		queryFn: () => penjualanApi.list(tanggal()),
		enabled: browser
	}));
}

/**
 * The omzet of one store-local day, with its breakdown by Pembayaran method and
 * by Kasir (#9). It is a query of its own rather than a sum over the sales list:
 * the aggregation is the API's to compute, and the list carries no Item lines to
 * compute it from anyway.
 */
export function createOmzetHarianQuery(tanggal: () => TanggalLaporan) {
	return createQuery<OmzetHarian, AppError>(() => ({
		queryKey: penjualanKeys.omzet(tanggal()),
		queryFn: () => penjualanApi.omzetHarian(tanggal()),
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

	return createMutation<HasilCheckout, AppError, CheckoutInput>(() => ({
		mutationFn: (input: CheckoutInput) => penjualanApi.checkout(input),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: produkKeys.all });
			queryClient.invalidateQueries({ queryKey: penjualanKeys.all });
		}
	}));
}

/**
 * Prints the Struk of one stored Penjualan again: what the Kasir presses when the
 * automatic print failed, and what `/penjualan` offers for a sale that already
 * happened. It is the same endpoint the checkout's automatic print goes through,
 * so both answer the same `{printed, message}` (ADR-0017, keputusan 5).
 *
 * Nothing is invalidated: printing changes no server state the app caches. The
 * Penjualan and the catalogue are exactly as they were, and the print result is
 * the answer the caller shows.
 */
export function createCetakStrukMutation() {
	return createMutation<HasilCetak, AppError, NomorStruk>(() => ({
		mutationFn: (nomorStruk: NomorStruk) => penjualanApi.cetakStruk(nomorStruk)
	}));
}

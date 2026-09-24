import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * Printing the Struk of one stored Penjualan again, proxied to Go with the session
 * token the browser cannot see (ADR-0001, ADR-0006).
 *
 * There is no Peran guard here and none in Go: CONTEXT.md gives the Kasir "cetak
 * Struk", so this sits behind the token check alone — the same as the checkout and
 * the sale read.
 *
 * The path segment is passed through untouched. Whether it names a Penjualan at
 * all is Go's to answer, and a print that failed comes back as the app's
 * `{printed: false, message}` rather than as an error status: the Penjualan is
 * stored either way (ADR-0017, keputusan 1).
 */
export const POST: RequestHandler = async (event) =>
	forward(
		await callBackend(event, `/api/penjualan/${event.params.receiptNumber}/struk`, {
			method: 'POST'
		})
	);

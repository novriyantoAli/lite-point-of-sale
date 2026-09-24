import { callBackend, forward } from '$lib/server/backend';
import type { RequestHandler } from './$types';

/**
 * Reading one stored Penjualan by its Nomor Struk, proxied to Go with the session
 * token the browser cannot see (ADR-0001, ADR-0006).
 *
 * There is no Peran guard here and none in Go: CONTEXT.md gives the Kasir "cetak
 * Struk", and both Peran look a Penjualan up at a one-terminal store, so this is
 * behind the token check alone.
 *
 * The path segment is passed through untouched — whether it names a Penjualan at
 * all is Go's to answer, and its 404 or 400 keeps the message Go gave it.
 */
export const GET: RequestHandler = async (event) =>
	forward(
		await callBackend(event, `/api/penjualan/${event.params.receiptNumber}`, { method: 'GET' })
	);

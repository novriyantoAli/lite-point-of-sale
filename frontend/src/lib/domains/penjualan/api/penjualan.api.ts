import { apiClient } from '$lib/api/client';
import {
	CetakEnvelopeSchema,
	CheckoutEnvelopeSchema,
	CheckoutInputSchema,
	NomorStrukSchema,
	OmzetHarianEnvelopeSchema,
	PenjualanEnvelopeSchema,
	PenjualanListSchema,
	TanggalLaporanSchema,
	type CheckoutInput,
	type HasilCetak,
	type HasilCheckout,
	type NomorStruk,
	type OmzetHarian,
	type Penjualan,
	type PenjualanRingkas,
	type TanggalLaporan
} from '../schemas/penjualan.schema';

/**
 * Interface next to implementation: `queries/` and tests depend on this shape,
 * not on the concrete object, so a fake can be injected without touching the
 * query layer (ADR-0007).
 *
 * Every call goes to the SvelteKit BFF on this origin — never to Go. The session
 * cookie rides along by itself, which is the whole point of Pattern A (ADR-0001,
 * ADR-0006).
 */
export interface PenjualanApi {
	/**
	 * Checks a cart out into a finished Penjualan. The API prices the Items from
	 * the catalogue, refuses a cart the Stok cannot cover, and works out the
	 * Kembalian — none of which this side sends or decides.
	 *
	 * It also prints the Struk, so the answer carries the sale *and* the outcome of
	 * that print. A print that failed is reported, not turned into a failed
	 * checkout (ADR-0017, keputusan 1).
	 */
	checkout(input: CheckoutInput): Promise<HasilCheckout>;
	/**
	 * One stored Penjualan, read by its Nomor Struk — what a reprint or a look-up
	 * arrives with. Whether the number names a Penjualan is Go's answer: a number
	 * that names nothing comes back as a normalized 404 ("Penjualan tidak
	 * ditemukan."), not as an empty Penjualan.
	 */
	getByReceiptNumber(nomorStruk: NomorStruk): Promise<Penjualan>;
	/**
	 * Prints the Struk of one stored Penjualan again, and answers whether the paper
	 * came out. It is the same use case the checkout runs automatically, so a Kasir
	 * whose first print failed retries through this one (ADR-0017, keputusan 5).
	 */
	cetakStruk(nomorStruk: NomorStruk): Promise<HasilCetak>;
	/**
	 * The Penjualan of one store-local day, newest Nomor Struk first: the sales list
	 * of #9. Each row is a summary — opening one reads the full Penjualan by its
	 * Nomor Struk through `getByReceiptNumber`.
	 */
	list(tanggal: TanggalLaporan): Promise<PenjualanRingkas[]>;
	/**
	 * The omzet of one store-local day: the total, the number of Penjualan, and the
	 * breakdown by Pembayaran method and by Kasir (#9).
	 */
	omzetHarian(tanggal: TanggalLaporan): Promise<OmzetHarian>;
}

export const penjualanApi: PenjualanApi = {
	async checkout(input: CheckoutInput): Promise<HasilCheckout> {
		const { data } = await apiClient.post('/penjualan', CheckoutInputSchema.parse(input));

		return CheckoutEnvelopeSchema.parse(data).data;
	},

	async getByReceiptNumber(nomorStruk: NomorStruk): Promise<Penjualan> {
		const { data } = await apiClient.get(`/penjualan/${NomorStrukSchema.parse(nomorStruk)}`);

		return PenjualanEnvelopeSchema.parse(data).data.sale;
	},

	async cetakStruk(nomorStruk: NomorStruk): Promise<HasilCetak> {
		const { data } = await apiClient.post(`/penjualan/${NomorStrukSchema.parse(nomorStruk)}/struk`);

		return CetakEnvelopeSchema.parse(data).data.print;
	},

	async list(tanggal: TanggalLaporan): Promise<PenjualanRingkas[]> {
		const { data } = await apiClient.get('/penjualan', {
			params: { date: TanggalLaporanSchema.parse(tanggal) }
		});

		return PenjualanListSchema.parse(data).data;
	},

	async omzetHarian(tanggal: TanggalLaporan): Promise<OmzetHarian> {
		const { data } = await apiClient.get('/penjualan/omzet', {
			params: { date: TanggalLaporanSchema.parse(tanggal) }
		});

		return OmzetHarianEnvelopeSchema.parse(data).data;
	}
};

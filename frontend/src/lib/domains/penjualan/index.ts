// Public surface of the penjualan domain. `api/` and `state/` stay internal: the
// query factories and the schema types are the surface every domain offers, so a
// route or another domain reaches them through here rather than through
// `queries/` or `schemas/` directly (ADR-0006). The till's own components are
// inside the domain, so they keep importing the factory from `queries/`.
export { default as Kasir } from './components/Kasir.svelte';
export { default as LaporanHarian } from './components/LaporanHarian.svelte';
export { default as PencarianPenjualan } from './components/PencarianPenjualan.svelte';

export {
	penjualanKeys,
	createCheckoutMutation,
	createCetakStrukMutation,
	createPenjualanDetailQuery,
	createPenjualanListQuery,
	createOmzetHarianQuery
} from './queries/penjualan.queries';

export {
	CetakEnvelopeSchema,
	CheckoutEnvelopeSchema,
	CheckoutInputSchema,
	CheckoutItemSchema,
	HasilCetakSchema,
	ItemPenjualanSchema,
	JumlahBayarSchema,
	METODE_LABEL,
	METODE_URUT,
	MetodePembayaranSchema,
	NomorStrukSchema,
	OmzetHarianEnvelopeSchema,
	OmzetHarianSchema,
	PembayaranSchema,
	PenjualanEnvelopeSchema,
	PenjualanListSchema,
	PenjualanRingkasSchema,
	PenjualanSchema,
	RingkasanKasirSchema,
	RingkasanMetodeSchema,
	TanggalLaporanSchema,
	tanggalHariIni,
	type CheckoutInput,
	type CheckoutItem,
	type HasilCetak,
	type HasilCheckout,
	type ItemPenjualan,
	type MetodePembayaran,
	type NomorStruk,
	type OmzetHarian,
	type Pembayaran,
	type Penjualan,
	type PenjualanRingkas,
	type RingkasanKasir,
	type RingkasanMetode,
	type TanggalLaporan
} from './schemas/penjualan.schema';

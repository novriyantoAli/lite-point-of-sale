// Public surface of the penjualan domain. `api/` and `state/` stay internal: the
// query factories and the schema types are the surface every domain offers, so a
// route or another domain reaches them through here rather than through
// `queries/` or `schemas/` directly (ADR-0006). The till's own components are
// inside the domain, so they keep importing the factory from `queries/`.
export { default as Kasir } from './components/Kasir.svelte';
export { default as PencarianPenjualan } from './components/PencarianPenjualan.svelte';

export {
	penjualanKeys,
	createCheckoutMutation,
	createCetakStrukMutation,
	createPenjualanDetailQuery
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
	PembayaranSchema,
	PenjualanEnvelopeSchema,
	PenjualanSchema,
	type CheckoutInput,
	type CheckoutItem,
	type HasilCetak,
	type HasilCheckout,
	type ItemPenjualan,
	type MetodePembayaran,
	type NomorStruk,
	type Pembayaran,
	type Penjualan
} from './schemas/penjualan.schema';

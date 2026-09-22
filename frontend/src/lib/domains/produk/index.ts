// Public surface of the produk domain. `api/` and `state/` stay internal: the
// query factories are exported because routes and components are what call them
// (ADR-0006).
export { default as ProdukForm } from './components/ProdukForm.svelte';
export { default as ProdukList } from './components/ProdukList.svelte';

export {
	produkKeys,
	createProdukListQuery,
	createKategoriListQuery,
	createProdukMutation,
	createUpdateProdukMutation,
	createSetProdukActiveMutation,
	createDeleteProdukMutation
} from './queries/produk.queries';

export {
	ProdukSchema,
	ProdukInputSchema,
	ProdukFilterSchema,
	ProdukEnvelopeSchema,
	ProdukListSchema,
	KategoriListSchema,
	type Produk,
	type ProdukInput,
	type ProdukFilter
} from './schemas/produk.schema';

// Public surface of the produk domain. `api/` and `state/` stay internal: the
// query factories are exported because routes and components are what call them
// (ADR-0006).
export { default as ProdukForm } from './components/ProdukForm.svelte';
export { default as ProdukList } from './components/ProdukList.svelte';
export { default as StokList } from './components/StokList.svelte';

export {
	produkKeys,
	createProdukListQuery,
	createKategoriListQuery,
	createStokMenipisQuery,
	createProdukMutation,
	createUpdateProdukMutation,
	createSetProdukActiveMutation,
	createAddStokMutation,
	createDeleteProdukMutation
} from './queries/produk.queries';

export {
	ProdukSchema,
	CreateProdukInputSchema,
	UpdateProdukInputSchema,
	ProdukFilterSchema,
	ProdukEnvelopeSchema,
	ProdukListSchema,
	KategoriListSchema,
	SetActiveInputSchema,
	TambahStokInputSchema,
	StokMenipisSchema,
	type Produk,
	type CreateProdukInput,
	type UpdateProdukInput,
	type ProdukFilter,
	type SetActiveInput,
	type TambahStokInput,
	type StokMenipis
} from './schemas/produk.schema';

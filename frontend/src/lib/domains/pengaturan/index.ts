// Public surface of the pengaturan domain. `api/` stays internal: the query
// factories are exported because routes and components are what call them
// (ADR-0006).
export { default as PengaturanForm } from './components/PengaturanForm.svelte';

export {
	pengaturanKeys,
	createPengaturanQuery,
	createStoreNameQuery,
	createUpdatePengaturanMutation
} from './queries/pengaturan.queries';

export {
	PAPER_WIDTH_OPTIONS,
	PengaturanSchema,
	PengaturanEnvelopeSchema,
	StoreNameEnvelopeSchema,
	UpdatePengaturanInputSchema,
	type Pengaturan,
	type UpdatePengaturanInput
} from './schemas/pengaturan.schema';

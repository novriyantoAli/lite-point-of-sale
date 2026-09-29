import { apiClient } from '$lib/api/client';
import {
	PengaturanEnvelopeSchema,
	StoreNameEnvelopeSchema,
	UpdatePengaturanInputSchema,
	type Pengaturan,
	type UpdatePengaturanInput
} from '../schemas/pengaturan.schema';

/**
 * Interface next to implementation: `queries/` and tests depend on this shape,
 * not on the concrete object, so a fake can be injected without touching the
 * query layer (ADR-0007).
 *
 * Every call goes to the SvelteKit BFF on this origin — never to Go. The session
 * cookie rides along by itself, which is the whole point of Pattern A (ADR-0001,
 * ADR-0006).
 */
export interface PengaturanApi {
	get(): Promise<Pengaturan>;
	/** The store's name, the one value readable before login (ADR-0019). */
	storeName(): Promise<string>;
	update(input: UpdatePengaturanInput): Promise<Pengaturan>;
}

export const pengaturanApi: PengaturanApi = {
	async get(): Promise<Pengaturan> {
		const { data } = await apiClient.get('/pengaturan');

		return PengaturanEnvelopeSchema.parse(data).data.settings;
	},

	async storeName(): Promise<string> {
		const { data } = await apiClient.get('/store-name');

		return StoreNameEnvelopeSchema.parse(data).data.store_name;
	},

	async update(input: UpdatePengaturanInput): Promise<Pengaturan> {
		const { data } = await apiClient.put('/pengaturan', UpdatePengaturanInputSchema.parse(input));

		return PengaturanEnvelopeSchema.parse(data).data.settings;
	}
};

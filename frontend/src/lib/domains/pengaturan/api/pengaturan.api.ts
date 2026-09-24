import { apiClient } from '$lib/api/client';
import {
	PengaturanEnvelopeSchema,
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
	update(input: UpdatePengaturanInput): Promise<Pengaturan>;
}

export const pengaturanApi: PengaturanApi = {
	async get(): Promise<Pengaturan> {
		const { data } = await apiClient.get('/pengaturan');

		return PengaturanEnvelopeSchema.parse(data).data.settings;
	},

	async update(input: UpdatePengaturanInput): Promise<Pengaturan> {
		const { data } = await apiClient.put('/pengaturan', UpdatePengaturanInputSchema.parse(input));

		return PengaturanEnvelopeSchema.parse(data).data.settings;
	}
};

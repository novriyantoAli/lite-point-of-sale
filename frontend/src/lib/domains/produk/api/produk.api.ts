import { apiClient } from '$lib/api/client';
import {
	KategoriListSchema,
	ProdukEnvelopeSchema,
	ProdukFilterSchema,
	ProdukInputSchema,
	ProdukListSchema,
	type Produk,
	type ProdukFilter,
	type ProdukInput
} from '../schemas/produk.schema';

/**
 * Interface next to implementation: `queries/` and tests depend on this shape,
 * not on the concrete object, so a fake can be injected without touching the
 * query layer (ADR-0007).
 *
 * Every call goes to the SvelteKit BFF on this origin — never to Go. The session
 * cookie rides along by itself, which is the whole point of Pattern A (ADR-0001,
 * ADR-0006).
 */
export interface ProdukApi {
	list(filter: ProdukFilter): Promise<Produk[]>;
	categories(): Promise<string[]>;
	create(input: ProdukInput): Promise<Produk>;
	update(id: number, input: ProdukInput): Promise<Produk>;
	setActive(id: number, active: boolean): Promise<Produk>;
	remove(id: number): Promise<void>;
}

export const produkApi: ProdukApi = {
	async list(filter: ProdukFilter): Promise<Produk[]> {
		const parsed = ProdukFilterSchema.parse(filter);
		const { data } = await apiClient.get('/produk', { params: queryParams(parsed) });

		return ProdukListSchema.parse(data).data;
	},

	async categories(): Promise<string[]> {
		const { data } = await apiClient.get('/produk/kategori');

		return KategoriListSchema.parse(data).data;
	},

	async create(input: ProdukInput): Promise<Produk> {
		const { data } = await apiClient.post('/produk', ProdukInputSchema.parse(input));

		return ProdukEnvelopeSchema.parse(data).data.product;
	},

	async update(id: number, input: ProdukInput): Promise<Produk> {
		const { data } = await apiClient.put(`/produk/${id}`, ProdukInputSchema.parse(input));

		return ProdukEnvelopeSchema.parse(data).data.product;
	},

	async setActive(id: number, active: boolean): Promise<Produk> {
		const { data } = await apiClient.patch(`/produk/${id}`, { active });

		return ProdukEnvelopeSchema.parse(data).data.product;
	},

	async remove(id: number): Promise<void> {
		await apiClient.delete(`/produk/${id}`);
	}
};

/**
 * Only the filters that were actually asked for reach the query string. Go reads
 * an empty one as "no filter" too, but leaving it off keeps the URL honest about
 * what was requested.
 */
function queryParams(filter: ProdukFilter): Record<string, string> {
	const params: Record<string, string> = {};

	if (filter.name) {
		params.name = filter.name;
	}
	if (filter.code) {
		params.code = filter.code;
	}
	if (filter.category) {
		params.category = filter.category;
	}
	if (filter.active !== undefined) {
		params.active = String(filter.active);
	}

	return params;
}

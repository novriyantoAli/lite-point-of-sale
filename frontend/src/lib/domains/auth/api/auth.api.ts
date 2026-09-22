import { apiClient } from '$lib/api/client';
import {
	CreatePenggunaInputSchema,
	LoginInputSchema,
	PenggunaListSchema,
	SessionSchema,
	type CreatePenggunaInput,
	type LoginInput,
	type Pengguna
} from '../schemas/auth.schema';

/**
 * Interface next to implementation: `queries/` and tests depend on this shape,
 * not on the concrete object, so a fake can be injected without touching the
 * query layer (ADR-0007).
 *
 * Every call goes to the SvelteKit BFF on this origin — never to Go. The
 * session cookie rides along by itself, which is the whole point of Pattern A
 * (ADR-0001, ADR-0006).
 */
export interface AuthApi {
	login(input: LoginInput): Promise<Pengguna>;
	logout(): Promise<void>;
	listPengguna(): Promise<Pengguna[]>;
	createPengguna(input: CreatePenggunaInput): Promise<Pengguna>;
	setPenggunaActive(id: number, active: boolean): Promise<Pengguna>;
}

export const authApi: AuthApi = {
	async login(input: LoginInput): Promise<Pengguna> {
		const { data } = await apiClient.post('/auth/login', LoginInputSchema.parse(input));

		return SessionSchema.parse(data).data.user;
	},

	async logout(): Promise<void> {
		await apiClient.post('/auth/logout');
	},

	async listPengguna(): Promise<Pengguna[]> {
		const { data } = await apiClient.get('/pengguna');

		return PenggunaListSchema.parse(data).data;
	},

	async createPengguna(input: CreatePenggunaInput): Promise<Pengguna> {
		const { data } = await apiClient.post('/pengguna', CreatePenggunaInputSchema.parse(input));

		return SessionSchema.parse(data).data.user;
	},

	async setPenggunaActive(id: number, active: boolean): Promise<Pengguna> {
		const { data } = await apiClient.patch(`/pengguna/${id}`, { active });

		return SessionSchema.parse(data).data.user;
	}
};

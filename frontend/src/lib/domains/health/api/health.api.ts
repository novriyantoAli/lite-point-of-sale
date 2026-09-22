import { apiClient } from '$lib/api/client';
import { HealthSchema, type Health } from '../schemas/health.schema';

/**
 * Interface next to implementation: `queries/` and tests depend on this shape,
 * not on the concrete object, so a fake can be injected without touching the
 * query layer (ADR-0007).
 */
export interface HealthApi {
	check(): Promise<Health>;
}

export const healthApi: HealthApi = {
	async check(): Promise<Health> {
		const { data } = await apiClient.get('/health');

		return HealthSchema.parse(data);
	}
};

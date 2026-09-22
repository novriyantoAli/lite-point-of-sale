import { browser } from '$app/environment';
import { createQuery } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import type { Health } from '../schemas/health.schema';
import { healthApi } from '../api/health.api';

export const healthKeys = {
	all: ['health'] as const,
	status: () => [...healthKeys.all, 'status'] as const
};

/**
 * Server state for the service health.
 *
 * `enabled: browser` keeps the fetch on the client: the request is same-origin
 * (`/api/...`), so it can only be made where the BFF is reachable, and the
 * dashboard does not need SSR.
 */
export function createHealthQuery() {
	return createQuery<Health, AppError>(() => ({
		queryKey: healthKeys.status(),
		queryFn: () => healthApi.check(),
		enabled: browser,
		refetchInterval: 30_000
	}));
}

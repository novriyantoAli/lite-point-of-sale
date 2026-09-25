import { browser } from '$app/environment';
import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
import type { AppError } from '$lib/api/errors';
import { backupApi } from '../api/backup.api';
import type { Backup } from '../schemas/backup.schema';

/**
 * The cache keys of the backup domain, in the one scheme every domain shares:
 * `["<domain>", "<shape>", ...]`.
 */
export const backupKeys = {
	all: ['backup'] as const,
	list: () => [...backupKeys.all, 'list'] as const
};

/**
 * The snapshot files on disk, oldest first. `enabled: browser` for the same
 * reason as the other domains: it is a same-origin call to the BFF, so it can
 * only be made where the BFF is reachable, and the screen does not need SSR.
 */
export function createBackupListQuery() {
	return createQuery<Backup[], AppError>(() => ({
		queryKey: backupKeys.list(),
		queryFn: () => backupApi.list(),
		enabled: browser
	}));
}

/**
 * Takes a manual snapshot. It invalidates the whole backup subtree so the list
 * the Admin just read grows by the one this call created.
 */
export function createBackupMutation() {
	const queryClient = useQueryClient();

	return createMutation<Backup, AppError, void>(() => ({
		mutationFn: () => backupApi.create(),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: backupKeys.all })
	}));
}

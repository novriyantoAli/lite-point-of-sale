import { apiClient } from '$lib/api/client';
import {
	BackupEnvelopeSchema,
	BackupListSchema,
	type Backup
} from '../schemas/backup.schema';

/**
 * Interface next to implementation: `queries/` and tests depend on this shape,
 * not on the concrete object, so a fake can be injected without touching the
 * query layer (ADR-0007).
 *
 * Every call goes to the SvelteKit BFF on this origin — never to Go. The session
 * cookie rides along by itself, which is the whole point of Pattern A (ADR-0001,
 * ADR-0006).
 */
export interface BackupApi {
	/** The snapshot files on disk, oldest first. */
	list(): Promise<Backup[]>;
	/** Takes a manual snapshot now and returns the file it created. */
	create(): Promise<Backup>;
}

export const backupApi: BackupApi = {
	async list(): Promise<Backup[]> {
		const { data } = await apiClient.get('/backup');

		return BackupListSchema.parse(data).data;
	},

	async create(): Promise<Backup> {
		const { data } = await apiClient.post('/backup');

		return BackupEnvelopeSchema.parse(data).data.backup;
	}
};

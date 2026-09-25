import { z } from 'zod';

/**
 * Contract of the backup domain, mirroring the Go DTOs of `adapter/httpapi`
 * 1:1 (ADR-0006): the schema is the only thing allowed to cross the HTTP
 * boundary, and it is the single source of the types below it.
 *
 * The field names are the API's, so they are English; the domain term, the route
 * (`/backup`) and the types here stay Indonesian — the boundary ADR-0012 records.
 */

/**
 * One snapshot of the store database as the API answers it. The filesystem path
 * is never sent: the browser learns a backup exists and how big it is, never
 * where on the server the file lives.
 */
export const BackupSchema = z.object({
	/** The file name of the snapshot within the backup folder. */
	name: z.string(),
	/** The snapshot length in bytes. */
	size: z.number().int().nonnegative(),
	/** When the snapshot was taken, RFC 3339 in UTC. */
	created_at: z.string()
});
export type Backup = z.infer<typeof BackupSchema>;

/** The answer to a manual export: one snapshot under `data.backup`. */
export const BackupEnvelopeSchema = z.object({
	data: z.object({ backup: BackupSchema })
});

/** The answer to the backup list: an array under `data`, oldest first. */
export const BackupListSchema = z.object({
	data: z.array(BackupSchema)
});

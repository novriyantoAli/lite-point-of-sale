// Public surface of the backup domain. `api/` stays internal: the query
// factories are exported because routes and components are what call them
// (ADR-0006).
export { default as BackupPanel } from './components/BackupPanel.svelte';

export { backupKeys, createBackupListQuery, createBackupMutation } from './queries/backup.queries';

export {
	BackupSchema,
	BackupEnvelopeSchema,
	BackupListSchema,
	type Backup
} from './schemas/backup.schema';

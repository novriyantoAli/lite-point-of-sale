import { readdir, rm } from 'node:fs/promises';
import path from 'node:path';

/**
 * Wipes the E2E database before the run.
 *
 * The Go API seeds its Admin Pengguna on the first start of a store and keeps
 * it there afterwards, so a file left over from an earlier run could hold
 * credentials this suite no longer knows — and the login spec would fail with a
 * message about a wrong password rather than about the leftover file. Every run
 * therefore starts from an empty store (ADR-0009).
 */
export default async function globalSetup() {
	const dataDir = path.resolve(import.meta.dirname, '../../backend/data');

	// WAL mode leaves -wal and -shm files next to the database; a stale one
	// confuses SQLite, so all three go.
	const entries = await readdir(dataDir).catch(() => [] as string[]);

	await Promise.all(
		entries
			.filter((name) => name.startsWith('e2e.db'))
			.map((name) => rm(path.join(dataDir, name), { force: true }))
	);
}

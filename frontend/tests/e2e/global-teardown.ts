import { rm } from 'node:fs/promises';
import type { FullConfig } from '@playwright/test';

/**
 * Removes the E2E store and printer file once the run is over.
 *
 * There is deliberately no matching setup that wipes them first: Playwright
 * starts `webServer` *before* global setup, so a wipe there deletes a file the
 * Go API already has open — the server keeps its unlinked inode, the suite runs
 * against the previous run's store, and the login specs fail with "username
 * sudah dipakai" instead of anything pointing at the cause. Each run therefore
 * uses files of its own (see playwright.config.ts) and this only tidies up after
 * it.
 */
export default async function globalTeardown(config: FullConfig) {
	const dbPath = config.metadata.e2eDbPath as string;
	const printerPath = config.metadata.e2ePrinterPath as string;

	// WAL mode leaves -wal and -shm beside the database; all three go. `force`
	// makes a missing file a no-op. The printer file is a plain file of ESC/POS
	// bytes.
	await Promise.all(
		[dbPath, `${dbPath}-wal`, `${dbPath}-shm`, printerPath].map((file) => rm(file, { force: true }))
	);
}

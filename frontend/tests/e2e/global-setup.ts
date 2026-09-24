import { writeFileSync } from 'node:fs';
import type { FullConfig } from '@playwright/test';

/**
 * Creates the file that stands in for the thermal printer.
 *
 * It runs once, in Playwright's own process, and that is the whole point: the
 * config module is loaded by every worker too, so creating the file there would
 * leave one orphan per worker — each worker has a pid of its own, and the path is
 * built from the pid. Playwright starts `webServer` before global setup, but the
 * Go API only opens the printer when a sale prints, which happens during the
 * tests — after this has run.
 *
 * The path comes from the config metadata, so the process that writes the bytes
 * and the process that reads them back agree on one file.
 */
export default async function globalSetup(config: FullConfig) {
	writeFileSync(config.metadata.e2ePrinterPath as string, '');
}

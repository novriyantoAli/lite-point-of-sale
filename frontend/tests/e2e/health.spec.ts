import { expect, test } from '@playwright/test';
import { logIn } from './helpers';

/**
 * The walking-skeleton happy path (ADR-0007, ADR-0009): a real browser loads the
 * dashboard, and the status it shows came from the Go API reading SQLite,
 * through the SvelteKit BFF. Nothing here is mocked — including the session, so
 * the browser logs in the way a Kasir would.
 */
test('dashboard reports the service healthy through the BFF', async ({ page }) => {
	await logIn(page);

	await expect(page.getByText('Status layanan')).toBeVisible();

	// The module is a ruled region of the mosaic, not a shadcn card: it is found by
	// the accessible name its own heading gives it, and the tag's word is read
	// exactly — `Basis data: ok` also contains "ok" (DESIGN.md, Do's: write every
	// state's word next to its mark).
	const status = page.getByRole('region', { name: 'Status layanan' });
	await expect(status.getByText('OK', { exact: true })).toBeVisible();
	await expect(status.getByText('Basis data: ok')).toBeVisible();
});

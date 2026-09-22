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

	// Scoped to the health card: the header carries a Badge of its own now.
	const card = page.locator('[data-slot="card"]').filter({ hasText: 'Status layanan' });
	await expect(card.locator('[data-slot="badge"]')).toHaveText('OK');
	await expect(card.getByText('Basis data: ok')).toBeVisible();
});

import { expect, test } from '@playwright/test';

/**
 * The walking-skeleton happy path (ADR-0007, ADR-0009): a real browser loads
 * the dashboard, and the status it shows came from the Go API reading SQLite,
 * through the SvelteKit BFF. Nothing here is mocked.
 */
test('dashboard reports the service healthy through the BFF', async ({ page }) => {
	await page.goto('/');

	await expect(page.getByText('Status layanan')).toBeVisible();
	await expect(page.locator('[data-slot="badge"]')).toHaveText('OK');
	await expect(page.getByText('Basis data: ok')).toBeVisible();
});

import { expect, test } from '@playwright/test';
import { createPengguna, logIn, logOut } from './helpers';

/**
 * The Backup happy paths (ADR-0007, ADR-0009): a real browser, the real
 * SvelteKit BFF, the real Go API on SQLite. Nothing is mocked.
 *
 * The file-level behaviour — that a snapshot really lands on disk as a valid
 * SQLite file — is proven at the REST seam in `backend/tests/e2e/backup_test.go`.
 * This spec proves the Admin's screen drives that flow and the Kasir's guard.
 */

test('the Admin takes a manual backup and it appears in the list', async ({ page }) => {
	await logIn(page);

	await page.goto('/backup');

	await page.getByRole('button', { name: 'Backup sekarang' }).click();

	// The screen reports the snapshot it just created.
	await expect(page.getByRole('status')).toContainText('Backup dibuat:');

	// The list shows at least one snapshot file (the daily backup may already have
	// seeded one at startup, so this asserts the list reflects a real file, not
	// that exactly one exists).
	await expect(page.locator('li').filter({ hasText: 'pos-' }).first()).toBeVisible();
});

test('a Kasir cannot open Backup, by link or by URL', async ({ page }) => {
	await logIn(page);
	await createPengguna(page, { username: 'kasir-backup-e2e', password: 'rahasia-kasir' });
	await logOut(page);

	await logIn(page, { username: 'kasir-backup-e2e', password: 'rahasia-kasir' });

	// The entry is hidden…
	await expect(page.getByRole('link', { name: 'Backup' })).toHaveCount(0);
	// …and typing the URL does not get around the guard: a backup is the Admin's
	// to take, and Go would refuse the action anyway.
	await page.goto('/backup');
	await expect(page).toHaveURL('/');
	await expect(page.getByRole('heading', { name: 'Backup' })).toHaveCount(0);

	await logOut(page);
});

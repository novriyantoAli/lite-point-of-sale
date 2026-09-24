import { expect, test } from '@playwright/test';
import { createPengguna, logIn, logOut } from './helpers';

/**
 * The Pengaturan happy paths (ADR-0007, ADR-0009): a real browser, the real
 * SvelteKit BFF, the real Go API on SQLite. Nothing is mocked.
 *
 * The ambang Stok menipis is deliberately *not* changed here: the suite shares
 * one store (ADR-0009), and the other specs read the seeded ambang of 5. Its
 * behaviour — that the stored value drives the restock list — is proven at the
 * REST seam in `backend/tests/e2e/pengaturan_test.go`.
 */

test('the Admin sets the Struk template and it persists', async ({ page }) => {
	await logIn(page);

	await page.goto('/pengaturan');

	// The ambang is seeded at 5; this test leaves it there, and asserts it shows.
	await expect(page.getByLabel('Ambang Stok menipis')).toHaveValue('5');

	// The two template blocks and the paper width are what "template" means here.
	await page.getByLabel('Header Struk').fill('Toko Kopi Purnama\nJl. Melati 1');
	await page.getByLabel('Footer Struk').fill('Terima kasih sudah belanja');
	await page.getByLabel('Lebar kertas').click();
	await page.getByRole('option', { name: '58 mm' }).click();

	await page.getByRole('button', { name: 'Simpan Pengaturan' }).click();

	await expect(page.getByRole('status')).toContainText('Pengaturan disimpan.');

	// The values are the stored ones: a reload reads them back.
	await page.reload();
	await expect(page.getByLabel('Header Struk')).toHaveValue('Toko Kopi Purnama\nJl. Melati 1');
	await expect(page.getByLabel('Footer Struk')).toHaveValue('Terima kasih sudah belanja');
	await expect(page.getByLabel('Lebar kertas')).toHaveText('58 mm');
	await expect(page.getByLabel('Ambang Stok menipis')).toHaveValue('5');
});

test('a Kasir cannot open Pengaturan, by link or by URL', async ({ page }) => {
	await logIn(page);
	await createPengguna(page, { username: 'kasir-pengaturan-e2e', password: 'rahasia-kasir' });
	await logOut(page);

	await logIn(page, { username: 'kasir-pengaturan-e2e', password: 'rahasia-kasir' });

	// The entry is hidden…
	await expect(page.getByRole('link', { name: 'Pengaturan' })).toHaveCount(0);
	// …and typing the URL does not get around the guard either: the settings are
	// the Admin's to change, like the catalogue.
	await page.goto('/pengaturan');
	await expect(page).toHaveURL('/');
	await expect(page.getByRole('heading', { name: 'Pengaturan' })).toHaveCount(0);

	await logOut(page);
});

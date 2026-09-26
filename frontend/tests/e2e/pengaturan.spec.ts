import { expect, test } from '@playwright/test';
import {
	bayarTunai,
	bukaKasir,
	bukaPenjualan,
	cariPenjualan,
	createPengguna,
	createProduk,
	logIn,
	logOut,
	nomorStruk,
	penjualanTersimpan,
	printedJob,
	printedSince,
	printedText,
	printerSize,
	tambahProduk
} from './helpers';

/**
 * The Pengaturan happy paths (ADR-0007, ADR-0009): a real browser, the real
 * SvelteKit BFF, the real Go API on SQLite. Nothing is mocked.
 *
 * The ambang Stok menipis is deliberately *not* changed here: the suite shares
 * one store (ADR-0009), and the other specs read the seeded ambang of 5. Its
 * behaviour — that the stored value drives the restock list — is proven at the
 * REST seam in `backend/tests/e2e/pengaturan_test.go`.
 *
 * The template *is* changed here, and the print that follows from it is asserted
 * here too: the Pengaturan row is one row for the whole store, so a second spec
 * rewriting it while this file saves-then-reads it would flake.
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

	// The first non-empty line of the header block *is* the store's name, and the
	// rail leads with it on every screen (PRODUCT.md, ADR-0019) — so the store
	// renaming itself here has to show up in the chrome, not only in the form.
	await expect(page.getByRole('banner').getByText('Toko Kopi Purnama')).toBeVisible();

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

test('changing the template changes a reprint of an old sale', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Template E2E', price: 4000, stock: 5 });

	await bukaKasir(page);
	await tambahProduk(page, 'Template E2E');
	await bayarTunai(page, 4000);
	const nomor = await nomorStruk(page);

	// The shop changes its Struk template the way an Admin does: through the form.
	await page.goto('/pengaturan');
	await page
		.getByLabel('Header Struk')
		.fill('Toko Baru E2E\nJl. Melati Nomor Dua Kelurahan Sukamaju');
	await page.getByLabel('Footer Struk').fill('Terima kasih E2E');
	await page.getByLabel('Lebar kertas').click();
	await page.getByRole('option', { name: '58 mm' }).click();
	await page.getByRole('button', { name: 'Simpan Pengaturan' }).click();
	await expect(page.getByRole('status')).toContainText('Pengaturan disimpan.');

	// Reprinting the sale that happened *before* the change prints the template as
	// it stands now, not a copy from when the sale happened (ADR-0017, keputusan 5).
	await bukaPenjualan(page);
	await cariPenjualan(page, String(nomor));

	const record = penjualanTersimpan(page);
	await expect(record.getByText(new RegExp(`Nomor Struk ${nomor}`))).toBeVisible();

	const before = printerSize();
	await record.getByRole('button', { name: 'Cetak Struk' }).click();
	await expect(record.getByText('Struk tercetak.')).toBeVisible();

	const printed = printedJob(printedSince(before), nomor);

	expect(printed).toContain('Toko Baru E2E');
	expect(printed).toContain('Terima kasih E2E');
	expect(printed).toContain(`Nomor Struk: ${nomor}`);

	// 58 mm means 32 columns: the long header line the Admin typed is wrapped, not
	// run past the roll. The job's control bytes are taken out first, so the check
	// is about what the paper holds.
	const lines = printedText(printed);
	expect(lines).toContain('Jl. Melati Nomor Dua Kelurahan');
	expect(lines).toContain('Sukamaju');
	for (const line of lines) {
		expect([...line].length).toBeLessThanOrEqual(32);
	}
});

import { readFileSync, statSync } from 'node:fs';
import { expect, test } from '@playwright/test';
import {
	bayarTunai,
	bukaKasir,
	bukaPenjualan,
	cariPenjualan,
	createProduk,
	logIn,
	nomorStruk,
	penjualanTersimpan,
	strukPenjualan,
	tambahProduk
} from './helpers';

/**
 * The Struk happy paths (ADR-0007, ADR-0009): a real browser, the real SvelteKit
 * BFF, the real Go API on SQLite, and a real ESC/POS print through the device
 * adapter — to a file that stands in for the thermal printer.
 *
 * `POS_PRINTER_DEVICE` is set to that file by playwright.config.ts, so the print
 * that runs automatically at checkout really writes bytes; reading them back here
 * is what proves the content of a Struk at the browser seam rather than only in
 * the encoder's own tests (ADR-0017).
 *
 * The file is shared by the whole run, so each assertion reads only the bytes
 * appended since a marker it took itself. Receipt numbers are never asserted to
 * be a particular value: the suite shares one store.
 */

/** The printer file this run writes to. */
function printerPath(): string {
	return test.info().config.metadata.e2ePrinterPath as string;
}

/** How much the printer file holds right now — the marker for "what comes next". */
function printerSize(): number {
	return statSync(printerPath()).size;
}

/** The bytes the printer file received after `from`, as text. */
function printedSince(from: number): string {
	return readFileSync(printerPath()).subarray(from).toString('utf8');
}

test('a checkout prints the Struk, and the Kasir sees that it did', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Cetak E2E', price: 5000, stock: 5 });

	await bukaKasir(page);
	await tambahProduk(page, 'Cetak E2E');

	// Everything the printer receives from here on is this sale's Struk.
	const before = printerSize();

	await bayarTunai(page, 10000);
	const nomor = await nomorStruk(page);

	// The panel reports the automatic print, and offers a reprint.
	await expect(strukPenjualan(page).getByText('Struk tercetak.')).toBeVisible();
	await expect(strukPenjualan(page).getByRole('button', { name: 'Cetak Struk' })).toBeVisible();

	// What actually reached the printer: the ESC/POS job, with the content
	// CONTEXT.md lists (ADR-0017, keputusan 4).
	const printed = printedSince(before);

	expect(printed).toContain('\x1b@');
	expect(printed).toContain(`Nomor Struk: ${nomor}`);
	expect(printed).toContain('Kasir: admin');
	expect(printed).toContain('Cetak E2E');
	expect(printed).toContain('  1 x Rp 5.000');
	expect(printed).toContain('Total');
	expect(printed).toContain('Bayar (Tunai)');
	expect(printed).toContain('Rp 10.000');
	expect(printed).toContain('Kembalian');
	expect(printed).toContain('Rp 5.000');
});

test('the Kasir reprints a Struk from the /penjualan lookup', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Cetak Ulang E2E', price: 7000, stock: 5 });

	await bukaKasir(page);
	await tambahProduk(page, 'Cetak Ulang E2E');
	await bayarTunai(page, 7000);
	const nomor = await nomorStruk(page);

	await bukaPenjualan(page);
	await cariPenjualan(page, String(nomor));

	const record = penjualanTersimpan(page);
	await expect(record.getByText(new RegExp(`Nomor Struk ${nomor}`))).toBeVisible();

	const before = printerSize();
	await record.getByRole('button', { name: 'Cetak Struk' }).click();

	await expect(record.getByText('Struk tercetak.')).toBeVisible();

	// The reprint goes through POST /api/penjualan/{nomorStruk}/struk, and the
	// bytes of the sale come out again (ADR-0017, keputusan 5).
	const printed = printedSince(before);

	expect(printed).toContain(`Nomor Struk: ${nomor}`);
	expect(printed).toContain('Cetak Ulang E2E');
	expect(printed).toContain('Total');
	expect(printed).toContain('Rp 7.000');
});

test('changing the template in Pengaturan changes a reprint of an old sale', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Template E2E', price: 4000, stock: 5 });

	await bukaKasir(page);
	await tambahProduk(page, 'Template E2E');
	await bayarTunai(page, 4000);
	const nomor = await nomorStruk(page);

	// The shop changes its Struk template the way an Admin does: through the form.
	await page.goto('/pengaturan');
	await page.getByLabel('Header Struk').fill('Toko Baru E2E\nJl. Melati 2');
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

	const printed = printedSince(before);

	expect(printed).toContain('Toko Baru E2E');
	expect(printed).toContain('Jl. Melati 2');
	expect(printed).toContain('Terima kasih E2E');
	expect(printed).toContain(`Nomor Struk: ${nomor}`);

	// 58 mm means 32 columns: the header the Admin just typed is wrapped to that,
	// never run past the roll.
	for (const line of printed.split('\n')) {
		expect([...line].length).toBeLessThanOrEqual(32);
	}
});

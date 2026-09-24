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
	printedJob,
	printedSince,
	printerSize,
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
 * The file is shared by the whole run, so each assertion reads the bytes appended
 * since a marker it took itself, and then narrows to its own sale's job — another
 * worker's Struk may have landed alongside it. Receipt numbers are never asserted
 * to be a particular value: the suite shares one store.
 *
 * Changing the template is deliberately *not* here: the Pengaturan row is one row
 * for the whole store, so a spec that rewrites it while `pengaturan.spec.ts`
 * saves-then-reads it would flake. That assertion lives in pengaturan.spec.ts,
 * the file that owns the form.
 */

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
	const printed = printedJob(printedSince(before), nomor);

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
	const printed = printedJob(printedSince(before), nomor);

	expect(printed).toContain(`Nomor Struk: ${nomor}`);
	expect(printed).toContain('Cetak Ulang E2E');
	expect(printed).toContain('Total');
	expect(printed).toContain('Rp 7.000');
});

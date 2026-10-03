import { expect, test, type Locator } from '@playwright/test';
import { createProduk, logIn, produkRow } from './helpers';

/**
 * Penjaga aturan angka DESIGN.md (Typography): "Semua angka uang, jumlah, dan
 * Stok memakai `font-variant-numeric: tabular-nums`, tanpa kecuali."
 *
 * Aturan itu tidak terlihat oleh tes teks, `getByRole`, maupun kontras — hanya
 * `getComputedStyle` yang membacanya. Sebelum spec ini ada, tidak satu pun
 * perintah §12 membacanya: kosakata tabel dulu disalin ke tiap layar, dan
 * `TD_ANGKA` di satu salinan kehilangan `tabular-nums` yang di salinan lain masih
 * ada, tanpa ada sinyal yang merah. Spec ini membaca aturannya kembali dari DOM
 * yang berjalan, seperti `permukaan-peramban.spec.ts` membaca permukaan lain.
 */

/**
 * Nilai `font-variant-numeric` sel angka pada kolom yang disebut namanya.
 *
 * Kolomnya dicari lewat `thead`, bukan indeks yang ditulis di tes: yang diuji
 * adalah selnya sendiri, bukan posisinya — jadi menambah kolom tidak membuat
 * penjaga ini menunjuk sel yang salah.
 */
async function angkaTabular(baris: Locator, kolom: string): Promise<string> {
	const tabel = baris.locator('xpath=ancestor::table[1]');
	const kepala = await tabel.locator('thead th').allTextContents();
	const indeks = kepala.findIndex((teks) => teks.trim() === kolom);

	return baris
		.locator('td')
		.nth(indeks)
		.evaluate((sel) => getComputedStyle(sel).fontVariantNumeric);
}

test('angka pada tabel memakai tabular-nums, tanpa kecuali', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Angka E2E', price: 18000, stock: 3 });

	// Produk: Harga dan Stok, dua kolom angka pada baris yang sama.
	await page.goto('/produk');
	const katalog = produkRow(page, 'Angka E2E');
	await expect(katalog).toBeVisible();
	expect(await angkaTabular(katalog, 'Harga')).toBe('tabular-nums');
	expect(await angkaTabular(katalog, 'Stok')).toBe('tabular-nums');

	// Stok: kolom Stok di tabel Stok per Produk.
	await page.goto('/stok');
	const stok = produkRow(page, 'Angka E2E');
	await expect(stok).toBeVisible();
	expect(await angkaTabular(stok, 'Stok')).toBe('tabular-nums');

	// Laporan: Omzet per metode Pembayaran, kolom Omzet — `—` atau angka, dan
	// keduanya berdiri di kolom angka.
	await page.goto('/laporan');
	const omzet = page
		.getByRole('table', { name: 'Omzet per metode Pembayaran' })
		.locator('tbody tr')
		.first();
	await expect(omzet).toBeVisible();
	expect(await angkaTabular(omzet, 'Omzet')).toBe('tabular-nums');
});

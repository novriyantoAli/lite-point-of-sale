import { expect, test } from '@playwright/test';

/**
 * Penjaga bahasa dokumen (PRODUCT.md: seluruh permukaan berbahasa Indonesia).
 *
 * `lang` hanya hidup di elemen <html> yang dirender `app.html`: bukan teks, bukan
 * tata letak, dan bukan nama, jadi `getByRole`, `toBeVisible`, kontras, serta
 * pengukuran mana pun diam terhadapnya — dan tidak satu pun perintah §12 membaca
 * file itu. Karena itu spec ini membacanya kembali dari dokumen yang benar-benar
 * disajikan: tanpa penjaga ini, `lang="en"` bisa bertahan di repo tanpa satu pun
 * sinyal hijau yang membuktikan apa-apa.
 */
test('dokumen menyatakan bahasa Indonesia', async ({ page }) => {
	await page.goto('/login');

	await expect(page.locator('html')).toHaveAttribute('lang', 'id');
});

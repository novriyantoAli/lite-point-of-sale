import { expect, test, type Locator } from '@playwright/test';
import { logIn, produkFilter } from './helpers';

/**
 * Penjaga dua titik menumpuk dunia (ADR-0020): kisi field menumpuk di 900px, papan
 * Kasir di 1080px. Keduanya bukan bawaan Tailwind `sm:` (640px), dan yang satu bukan
 * yang lain — kisi saringan Produk pernah memakai titik papan, dan itulah yang tes ini
 * kunci.
 *
 * Dibaca dari `grid-template-columns` DOM yang berjalan: layout tidak terlihat oleh tes
 * teks, `getByRole`, maupun kontras. Hanya mengukur bisa membedakannya.
 */
function kolom(locator: Locator): Promise<number> {
	return locator.evaluate(
		(element) => getComputedStyle(element).gridTemplateColumns.split(' ').length
	);
}

test('kisi field menumpuk di 900px, bukan di titik papan 1080px', async ({ page }) => {
	await logIn(page);

	// Di antara kedua titik: papan sudah menumpuk, tetapi kisi field masih empat kolom.
	await page.setViewportSize({ width: 1000, height: 900 });
	await page.goto('/produk');

	// The field, its FIELD wrapper, then the form grid itself.
	const grid = produkFilter(page).getByLabel('Nama').locator('xpath=../..');
	await expect(grid).toBeVisible();
	expect(await kolom(grid)).toBe(4);

	// Di 900px ia menumpuk jadi satu kolom.
	await page.setViewportSize({ width: 900, height: 900 });
	expect(await kolom(grid)).toBe(1);
});

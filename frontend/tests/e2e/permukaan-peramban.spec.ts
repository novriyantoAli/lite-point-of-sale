import { expect, test, type Locator, type Page } from '@playwright/test';
import { bukaKasir, createProduk, logIn, pembayaran, produkFilter, produkRow } from './helpers';

/**
 * Penjaga empat permukaan peramban yang diwarnai DESIGN.md (Browser surfaces;
 * Do's: "theme the browser surfaces — selection, caret, scrollbar, focus ring").
 *
 * Keempatnya bukan teks, bukan tata letak, dan bukan nama: getByRole, toBeVisible,
 * kontras, radius, bayangan, dan tabular-nums semuanya diam terhadapnya. Karena itu
 * spec ini membaca properti itu kembali dari DOM aplikasi yang berjalan, dan merah
 * begitu aturannya hilang dari app.css. Warnanya dibandingkan dengan token dunia ini,
 * bukan literal `rgb(...)` — token boleh berubah, aturannya tidak.
 */

/** Cincin fokus dunia ini, sama di setiap permukaan: 2px ke dalam, bukan glow. */
const RING = { width: '2px', style: 'solid', offset: '-2px' };

/**
 * Nilai satu token `:root` dalam bentuk yang sama dengan `getComputedStyle`
 * mengembalikannya. Sebuah elemen probe menyelesaikan `var(...)`-nya, jadi
 * `--primary-foreground: #ffffff` dan `outline-color: rgb(255, 255, 255)` bisa
 * dibandingkan tanpa menuliskan literalnya di tes.
 */
async function tokenColor(page: Page, token: string): Promise<string> {
	return page.evaluate((name) => {
		const probe = document.createElement('span');
		probe.style.color = `var(${name})`;
		document.body.appendChild(probe);
		const resolved = getComputedStyle(probe).color;
		probe.remove();
		return resolved;
	}, token);
}

/** Memfokuskan seperti pengguna papan ketik, supaya `:focus-visible` yang cocok. */
async function focusVisible(locator: Locator) {
	await locator.evaluate((element: HTMLElement) => {
		// Blur dulu: kalau elemennya sudah memegang fokus (dialog mengembalikannya ke
		// tombol pemicu), `focus()` jadi no-op dan `focusVisible` tak pernah menyala.
		(document.activeElement as HTMLElement | null)?.blur();
		element.focus({ focusVisible: true } as FocusOptions);
	});
	await expect(locator).toBeFocused();
}

/**
 * Properti cincin fokus yang dibaca sekaligus, supaya perbandingannya utuh.
 *
 * Dengan `{ focus: true }` fokus dan pembacaannya terjadi di satu evaluasi yang sama:
 * pembacaan biasa bisa mendarat setelah transisi 150ms primitif selesai, dan penjaga
 * "mendarat seketika" lolos tanpa membuktikan apa pun.
 */
async function ringOf(locator: Locator, options: { focus?: boolean } = {}) {
	return locator.evaluate((element: HTMLElement, opts: { focus?: boolean }) => {
		if (opts.focus) {
			(document.activeElement as HTMLElement | null)?.blur();
			element.focus({ focusVisible: true } as FocusOptions);
		}
		const style = getComputedStyle(element);
		return {
			color: style.outlineColor,
			width: style.outlineWidth,
			style: style.outlineStyle,
			offset: style.outlineOffset
		};
	}, options);
}

/** Setiap durasi dalam daftar bernilai nol — `prefers-reduced-motion` menolkan semuanya. */
function isZeroDurations(value: string): boolean {
	return value
		.split(',')
		.map((part) => part.trim())
		.every((part) => part === '0s');
}

test('teks terpilih memakai tinta penuh dengan teks petak', async ({ page }) => {
	await logIn(page);

	const ink = await tokenColor(page, '--foreground');
	const tile = await tokenColor(page, '--card');

	const selection = await page.evaluate(() => {
		const style = getComputedStyle(document.documentElement, '::selection');
		return { background: style.backgroundColor, color: style.color };
	});

	expect(selection.background).toBe(ink);
	expect(selection.color).toBe(tile);
});

test('caret dan scrollbar mengikuti palet', async ({ page }) => {
	await logIn(page);

	const ink = await tokenColor(page, '--foreground');
	const hairline = await tokenColor(page, '--border');
	const ground = await tokenColor(page, '--background');

	const html = await page.evaluate(() => {
		const style = getComputedStyle(document.documentElement);
		return {
			caret: style.caretColor,
			scrollbarWidth: style.scrollbarWidth,
			scrollbarColor: style.scrollbarColor
		};
	});

	expect(html.caret).toBe(ink);
	expect(html.scrollbarWidth).toBe('thin');
	expect(html.scrollbarColor).toBe(`${hairline} ${ground}`);

	// Satu-satunya caret yang disentuh merah utilitas: field jumlah bayar Kasir,
	// lewat `data-caret="utility"` (bukan id, supaya app.css tak tahu id domain).
	await bukaKasir(page);
	const red = await tokenColor(page, '--destructive');
	const bayar = await pembayaran(page)
		.getByRole('textbox', { name: 'Jumlah bayar' })
		.evaluate((element) => getComputedStyle(element).caretColor);

	expect(bayar).toBe(red);
});

test('cincin fokus memakai tinta, dan petak di atas bidang bertinta', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Cincin E2E', price: 5000, stock: 1 });

	const ink = await tokenColor(page, '--foreground');
	const tile = await tokenColor(page, '--primary-foreground');

	// Bidang bertinta penuh (bg-primary): cincinnya petak, bukan tinta — tinta di
	// atas tinta terukur 1,00:1 (DESIGN.md, Browser surfaces).
	const primary = page.getByRole('button', { name: 'Tambah Produk' });
	await focusVisible(primary);
	expect(await ringOf(primary)).toEqual({ ...RING, color: tile });

	// Input: cincin tinta.
	const input = produkFilter(page).getByLabel('Nama');
	await focusVisible(input);
	expect(await ringOf(input)).toEqual({ ...RING, color: ink });

	// Aksi baris `variant="ghost"`: cincin tinta.
	const ghost = produkRow(page, 'Cincin E2E').getByRole('button', { name: 'Ubah' });
	await focusVisible(ghost);
	expect(await ringOf(ghost)).toEqual({ ...RING, color: ink });
});

test('cincin fokus mendarat seketika, bukan bertransisi', async ({ page }) => {
	await logIn(page);
	await page.goto('/produk');

	const tile = await tokenColor(page, '--primary-foreground');
	const primary = page.getByRole('button', { name: 'Tambah Produk' });

	// Dibaca di evaluasi yang sama dengan fokusnya (tepat saat cincin mendarat), lalu
	// lagi setelah transisi 150ms primitif seharusnya selesai; keduanya harus sama,
	// atau `transition-all` menganimasikan cincinnya (3px/offset 0 pada ~150ms pertama
	// — bug yang #41 rekam).
	const onLanding = await ringOf(primary, { focus: true });
	await page.waitForTimeout(250);
	const settled = await ringOf(primary);

	expect(onLanding).toEqual({ ...RING, color: tile });
	expect(settled).toEqual({ ...RING, color: tile });
	expect(onLanding).toEqual(settled);
});

test('cincin sel metode terpilih memakai petak', async ({ page }) => {
	await logIn(page);
	await bukaKasir(page);

	const tile = await tokenColor(page, '--primary-foreground');
	const tunai = page.getByRole('radio', { name: 'Tunai' });
	const cell = tunai.locator('..');

	// Sel metode melukis bidang bertintanya lewat `has-[:checked]`, bukan `bg-primary`,
	// jadi cincinnya ditanggung komponennya sendiri.
	await tunai.evaluate((element: HTMLElement) => element.focus());
	await expect(tunai).toBeFocused();

	expect(await ringOf(cell)).toEqual({ ...RING, color: tile });
});

test('prefers-reduced-motion menolkan durasi transisi dan animasi', async ({ page }) => {
	await page.emulateMedia({ reducedMotion: 'reduce' });
	await logIn(page);
	await page.goto('/produk');

	// Tombol: transisi dimatikan.
	const button = page.getByRole('button', { name: 'Tambah Produk' });
	expect(
		isZeroDurations(
			await button.evaluate((element) => getComputedStyle(element).transitionDuration)
		)
	).toBe(true);

	// Dialog: animasi masuknya (dari tw-animate-css) juga dimatikan.
	await button.click();
	const content = page.locator('[data-slot="dialog-content"]');
	await expect(content).toBeVisible();
	expect(
		isZeroDurations(
			await content.evaluate((element) => getComputedStyle(element).animationDuration)
		)
	).toBe(true);
});

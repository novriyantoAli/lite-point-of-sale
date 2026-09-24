import { expect, type Page } from '@playwright/test';

/**
 * The credentials of the seeded Admin Pengguna, matching the
 * POS_ADMIN_USERNAME/POS_ADMIN_PASSWORD the Go API is started with in
 * playwright.config.ts. Every run gets its own empty store (`e2e-<pid>.db`),
 * so the seed runs on every E2E run with exactly these credentials.
 */
export const ADMIN = { username: 'admin', password: 'rahasia-admin' };

/** Logs in through the real login form and waits for the dashboard. */
export async function logIn(
	page: Page,
	credentials: { username: string; password: string } = ADMIN
) {
	await page.goto('/login');
	await page.getByLabel('Username').fill(credentials.username);
	await page.getByLabel('Password').fill(credentials.password);
	await page.getByRole('button', { name: 'Masuk' }).click();

	await expect(page).toHaveURL('/');
	await expect(page.getByText(credentials.username, { exact: true })).toBeVisible();
}

/** Logs out through the header button and waits for the login page. */
export async function logOut(page: Page) {
	await page.getByRole('button', { name: 'Keluar' }).click();

	await expect(page).toHaveURL('/login');
}

/** The staff row of one Pengguna, so a test never matches a similar username. */
export function penggunaRow(page: Page, username: string) {
	return page.locator('li').filter({ has: page.getByText(username, { exact: true }) });
}

/**
 * Creates a Pengguna through the Admin UI. `role` is picked in the shadcn
 * Select only when it is not the default, which also keeps one test exercising
 * the dropdown.
 */
export async function createPengguna(
	page: Page,
	{
		username,
		password,
		role = 'Kasir'
	}: { username: string; password: string; role?: 'Kasir' | 'Admin' }
) {
	await page.goto('/pengguna');

	await page.getByLabel('Username').fill(username);
	await page.getByLabel('Password').fill(password);

	if (role !== 'Kasir') {
		await page.getByLabel('Peran').click();
		await page.getByRole('option', { name: role }).click();
	}

	await page.getByRole('button', { name: 'Tambah', exact: true }).click();

	// The form reports success and the new Pengguna shows up in the list.
	await expect(page.getByRole('status')).toContainText(username);
	await expect(penggunaRow(page, username)).toBeVisible();
}

/**
 * The catalogue filter bar. It shares its field labels with the Produk form
 * below it — deliberately, because that is what each is to the person using it
 * — so a test that types without scoping would be typing into the wrong one.
 */
export function produkFilter(page: Page) {
	return page.getByRole('search', { name: 'Saring katalog' });
}

/** The Produk form, scoped for the same reason as `produkFilter`. */
export function produkForm(page: Page) {
	return page.getByRole('form', { name: 'Formulir Produk' });
}

/** The catalogue row of one Produk, so a test never matches a similar name. */
export function produkRow(page: Page, name: string) {
	return page.locator('tbody tr').filter({ has: page.getByText(name, { exact: true }) });
}

/**
 * Adds a Produk through the Admin UI and waits for it to appear in the
 * catalogue. Kode and Kategori are only filled when a test needs them, which
 * keeps the "both may be left blank" rule exercised by the other tests.
 */
export async function createProduk(
	page: Page,
	{
		name,
		code,
		price,
		stock,
		category
	}: { name: string; code?: string; price: number; stock: number; category?: string }
) {
	await page.goto('/produk');
	await page.getByRole('button', { name: 'Tambah Produk' }).click();

	const form = produkForm(page);
	await form.getByLabel('Nama').fill(name);
	if (code !== undefined) {
		await form.getByLabel('Kode').fill(code);
	}
	await form.getByLabel('Harga').fill(String(price));
	await form.getByLabel('Stok').fill(String(stock));
	if (category !== undefined) {
		await form.getByLabel('Kategori').fill(category);
	}

	await page.getByRole('button', { name: 'Tambah', exact: true }).click();

	// The form reports success and the new Produk shows up in the catalogue.
	await expect(page.getByRole('status')).toContainText(name);
	await expect(produkRow(page, name)).toBeVisible();
}

/** Fills the Produk form without submitting it, for the refusal tests. */
export async function fillProdukForm(
	page: Page,
	{ name, code, price, stock }: { name: string; code?: string; price: number; stock: number }
) {
	const form = produkForm(page);
	await form.getByLabel('Nama').fill(name);
	if (code !== undefined) {
		await form.getByLabel('Kode').fill(code);
	}
	await form.getByLabel('Harga').fill(String(price));
	await form.getByLabel('Stok').fill(String(stock));
}

/**
 * The Stok screen's restock list. It and the catalogue table below it carry the
 * same Produk on purpose, so a test has to say which one it is reading —
 * otherwise "Tambah Stok" would match a button in each.
 */
export function stokMenipisSection(page: Page) {
	return page.getByRole('region', { name: 'Stok menipis' });
}

/** The Stok screen's table of every Produk. */
export function stokProdukSection(page: Page) {
	return page.getByRole('region', { name: 'Stok per Produk' });
}

/** The Stok row of one Produk, so a test never matches a similar name. */
export function stokRow(page: Page, name: string) {
	return stokProdukSection(page)
		.locator('tbody tr')
		.filter({ has: page.getByText(name, { exact: true }) });
}

/**
 * Records a restock through the Stok screen for one Produk and waits for the
 * screen to report it. The amount is the units received, never the new total —
 * the Stok already on the Produk is the API's to add to.
 */
export async function addStok(page: Page, name: string, quantity: number) {
	await page.goto('/stok');

	const row = stokRow(page, name);
	await row.getByRole('button', { name: 'Tambah Stok' }).click();
	await row.getByLabel('Jumlah masuk').fill(String(quantity));
	await row.getByRole('button', { name: 'Tambah Stok' }).click();

	await expect(page.getByRole('status')).toContainText(`Stok ${name} sekarang`);
}

/** The till's Produk lookup results. */
export function kasirHasil(page: Page) {
	return page.getByRole('list', { name: 'Hasil pencarian Produk' });
}

/**
 * The till's Kode field — what a barcode scanner types into. It is matched by
 * role and exactly, because `getByLabel` reads the `Tambah <nama> ke keranjang`
 * buttons' labels as a substring match too.
 */
export function kasirKode(page: Page) {
	return page.getByRole('textbox', { name: 'Kode', exact: true });
}

/** The till's name search field. */
export function kasirNama(page: Page) {
	return page.getByRole('textbox', { name: 'Nama', exact: true });
}

/** The keranjang section of the till. */
export function keranjang(page: Page) {
	return page.getByRole('region', { name: 'Keranjang' });
}

/** The quantity field of one keranjang line. */
export function jumlahItem(page: Page, name: string) {
	return keranjang(page).getByRole('spinbutton', { name: `Jumlah ${name}` });
}

/** The Pembayaran form of the till: the method, and Tunai's amount. */
export function pembayaran(page: Page) {
	return page.getByRole('form', { name: 'Pembayaran' });
}

/** The recorded-Penjualan panel that replaces the till once a sale is stored. */
export function strukPenjualan(page: Page) {
	return page.getByRole('region', { name: 'Penjualan tercatat' });
}

/** Opens the till. */
export async function bukaKasir(page: Page) {
	await page.goto('/kasir');
}

/** Adds a Produk from the till's lookup to the keranjang. */
export async function tambahProduk(page: Page, name: string) {
	await page.getByRole('button', { name: `Tambah ${name} ke keranjang` }).click();
}

/**
 * Pays the keranjang with a Tunai amount and waits for the recorded Penjualan.
 * The amount is what the buyer handed over, never the total — the Kembalian is
 * the API's to work out.
 */
export async function bayarTunai(page: Page, amount: number) {
	await pembayaran(page).getByRole('textbox', { name: 'Jumlah bayar' }).fill(String(amount));
	await pembayaran(page).getByRole('button', { name: 'Bayar & Simpan Penjualan' }).click();

	await expect(strukPenjualan(page)).toBeVisible();
}

/**
 * Pays the keranjang with a recorded method — QRIS, Debit or Transfer — and waits
 * for the recorded Penjualan. There is nothing to type: no gateway is called, and
 * the method pays the total (CONTEXT.md, Pembayaran).
 *
 * The method is picked by clicking its label, which is what a Kasir clicks: the
 * radio itself is hidden from sight and only carries the choice.
 */
export async function bayarNonTunai(page: Page, label: string) {
	// The pill is what a Kasir clicks; the radio inside it is hidden from sight and
	// only carries the choice. Clicking the label is also what proves the label
	// names the radio, which is how the method is announced.
	await pembayaran(page).getByText(label, { exact: true }).click();
	await expect(pembayaran(page).getByRole('radio', { name: label })).toBeChecked();

	await pembayaran(page).getByRole('button', { name: 'Bayar & Simpan Penjualan' }).click();

	await expect(strukPenjualan(page)).toBeVisible();
}

/** The Nomor Struk the recorded-Penjualan panel is showing. */
export async function nomorStruk(page: Page): Promise<number> {
	const text = await strukPenjualan(page)
		.getByText(/Nomor Struk/)
		.textContent();

	return Number(/Nomor Struk\s+(\d+)/.exec(text ?? '')?.[1]);
}

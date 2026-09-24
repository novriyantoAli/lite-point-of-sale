import { expect, test } from '@playwright/test';
import {
	bayarNonTunai,
	bayarTunai,
	bukaKasir,
	createPengguna,
	createProduk,
	jumlahItem,
	kasirHasil,
	kasirKode,
	kasirNama,
	keranjang,
	logIn,
	logOut,
	nomorStruk,
	pembayaran,
	produkRow,
	stokRow,
	strukPenjualan,
	tambahProduk
} from './helpers';

/**
 * The Penjualan Tunai happy paths (ADR-0007, ADR-0009): a real browser, the real
 * SvelteKit BFF, the real Go API on SQLite. Nothing is mocked — including the
 * Stok leaving the catalogue and the Nomor Struk being assigned, which are SQL
 * statements and can only be shown to work against a real database.
 *
 * Every test adds its own Produk under a name of its own, because the suite
 * shares one store for the whole run. Receipt numbers are never asserted to be a
 * particular value for the same reason: they are global and keep counting across
 * every sale the run makes.
 */

test('the Kasir sells a Produk and the Stok drops by what was sold', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Jual E2E', price: 5000, stock: 10 });

	await bukaKasir(page);
	await tambahProduk(page, 'Jual E2E');
	await keranjang(page).getByRole('button', { name: 'Tambah jumlah Jual E2E' }).click();

	// The running total follows the quantity: two units at 5000.
	await expect(keranjang(page).getByText('2 unit dalam 1 Item.')).toBeVisible();
	await expect(keranjang(page).getByRole('status')).toHaveText('Total Rp 10.000');

	await bayarTunai(page, 20000);

	// The Nomor Struk and the Kembalian are the API's numbers, not this screen's
	// arithmetic.
	await expect(strukPenjualan(page).getByText(/Nomor Struk/)).toContainText(/Nomor Struk \d+/);
	const kembalian = strukPenjualan(page).locator('p').filter({ hasText: 'Kembalian' });
	await expect(kembalian).toContainText('Rp 10.000');
	await expect(strukPenjualan(page).locator('p').filter({ hasText: 'Bayar' })).toContainText(
		'Rp 20.000'
	);

	// The Stok left the catalogue by exactly what was sold.
	await page.goto('/stok');
	await expect(stokRow(page, 'Jual E2E')).toContainText('8');
});

test('the Kasir finds a Produk by its Kode', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Kode E2E', code: 'KODE-E2E-01', price: 7000, stock: 5 });

	await bukaKasir(page);

	// What a barcode scanner types into the Kode field narrows the listing to the
	// one Produk that Kode belongs to.
	await kasirKode(page).fill('KODE-E2E-01');

	await expect(kasirHasil(page).getByText(/Kode E2E/)).toBeVisible();
	await expect(kasirHasil(page).getByText(/Jual E2E/)).toHaveCount(0);

	await tambahProduk(page, 'Kode E2E');

	await expect(jumlahItem(page, 'Kode E2E')).toHaveValue('1');
});

test('a Produk without a Kode is still found, by its name', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Nama Saja E2E', price: 4000, stock: 5 });

	await bukaKasir(page);
	await kasirNama(page).fill('Nama Saja');

	await expect(kasirHasil(page).getByText('Nama Saja E2E')).toBeVisible();

	await tambahProduk(page, 'Nama Saja E2E');

	await expect(jumlahItem(page, 'Nama Saja E2E')).toHaveValue('1');
});

test('the Kasir removes an Item, and empties the whole keranjang', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Keranjang Satu E2E', price: 5000, stock: 5 });
	await createProduk(page, { name: 'Keranjang Dua E2E', price: 6000, stock: 5 });

	await bukaKasir(page);
	await tambahProduk(page, 'Keranjang Satu E2E');
	await tambahProduk(page, 'Keranjang Dua E2E');

	await keranjang(page)
		.getByRole('button', { name: 'Hapus Keranjang Satu E2E dari keranjang' })
		.click();
	await expect(jumlahItem(page, 'Keranjang Satu E2E')).toHaveCount(0);
	await expect(jumlahItem(page, 'Keranjang Dua E2E')).toBeVisible();

	await keranjang(page).getByRole('button', { name: 'Kosongkan' }).click();

	await expect(keranjang(page).getByText(/Keranjang kosong/)).toBeVisible();
});

test('checkout is blocked while a line asks for more than the Stok', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Sedikit E2E', price: 5000, stock: 1 });

	await bukaKasir(page);
	await tambahProduk(page, 'Sedikit E2E');
	await keranjang(page).getByRole('button', { name: 'Tambah jumlah Sedikit E2E' }).click();

	await expect(keranjang(page).getByText(/Melebihi Stok: tersisa 1/)).toBeVisible();
	await expect(
		pembayaran(page).getByRole('button', { name: 'Bayar & Simpan Penjualan' })
	).toBeDisabled();
});

test('a payment below the total is refused, and nothing is sold', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Kurang Bayar E2E', price: 5000, stock: 3 });

	await bukaKasir(page);
	await tambahProduk(page, 'Kurang Bayar E2E');

	await pembayaran(page).getByRole('textbox', { name: 'Jumlah bayar' }).fill('1000');
	await pembayaran(page).getByRole('button', { name: 'Bayar & Simpan Penjualan' }).click();

	await expect(pembayaran(page).getByText('Jumlah bayar kurang dari total.')).toBeVisible();

	// Refused means refused: no Penjualan was recorded and the Stok is where it was.
	await expect(strukPenjualan(page)).toHaveCount(0);
	await page.goto('/stok');
	await expect(stokRow(page, 'Kurang Bayar E2E')).toContainText('3');
});

test('the Kasir takes a non-tunai Pembayaran, and the Struk names the method', async ({ page }) => {
	await logIn(page);

	// Each method gets a Produk of its own: the suite shares one store for the
	// whole run, and the Stok assertion below would otherwise depend on the order
	// the methods ran in.
	for (const metode of ['QRIS', 'Debit', 'Transfer']) {
		const name = `Non-tunai ${metode} E2E`;
		await createProduk(page, { name, price: 7000, stock: 5 });

		await bukaKasir(page);
		await tambahProduk(page, name);
		await bayarNonTunai(page, metode);

		// What the Struk shows: the method and the total, and no Kembalian at all —
		// a recorded method pays the total exactly (CONTEXT.md, Kembalian).
		const bayar = strukPenjualan(page).locator('p').filter({ hasText: 'Bayar' });
		await expect(bayar).toContainText(metode);
		await expect(bayar).toContainText('Rp 7.000');
		await expect(strukPenjualan(page).getByText('Kembalian')).toHaveCount(0);

		// The API stored it: the sale took the Stok out, so it was written rather
		// than just echoed. What the Pembayaran row holds is asserted at the REST
		// seam in backend/tests/e2e/penjualan_test.go — the BFF has no reader for a
		// Penjualan by its Nomor Struk yet (#8, #9 build one).
		await page.goto('/stok');
		await expect(stokRow(page, name)).toContainText('4');
	}
});

test('every Penjualan gets a Nomor Struk of its own that keeps counting', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Nomor E2E', price: 5000, stock: 5 });

	await bukaKasir(page);
	await tambahProduk(page, 'Nomor E2E');
	await bayarTunai(page, 5000);
	const first = await nomorStruk(page);

	// A new sale from the same till, without leaving the screen.
	await page.getByRole('button', { name: 'Penjualan Baru' }).click();
	await tambahProduk(page, 'Nomor E2E');
	await bayarTunai(page, 5000);
	const second = await nomorStruk(page);

	// Global and never reset — and never reused, which is why the second one is
	// not the first one again.
	expect(second).toBeGreaterThan(first);
});

test('a Kasir sells from the till, and the Struk names them', async ({ page }) => {
	await logIn(page);
	await createPengguna(page, { username: 'kasir-jual-e2e', password: 'rahasia-kasir' });
	await createProduk(page, { name: 'Kasir Jual E2E', price: 4000, stock: 5 });
	await logOut(page);

	await logIn(page, { username: 'kasir-jual-e2e', password: 'rahasia-kasir' });

	// The till is the one catalogue-adjacent screen a Kasir gets; the Admin screens
	// stay out of reach.
	await expect(page.getByRole('link', { name: 'Kasir' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Produk' })).toHaveCount(0);

	await bukaKasir(page);
	await tambahProduk(page, 'Kasir Jual E2E');
	await bayarTunai(page, 4000);

	await expect(strukPenjualan(page).getByText(/Kasir kasir-jual-e2e/)).toBeVisible();

	await logOut(page);
});

test('a Produk that has been sold can only be deactivated, never deleted', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Terjual E2E', price: 5000, stock: 3 });

	// A finished Penjualan is the only thing that makes a Produk "pernah terjual"
	// (#6 — the gap #4 left).
	await bukaKasir(page);
	await tambahProduk(page, 'Terjual E2E');
	await bayarTunai(page, 5000);

	// The API refuses the delete, through the real BFF and the real Go…
	const listed = await page.request.get('/api/produk?name=Terjual%20E2E');
	const [produk] = (await listed.json()).data as { id: number }[];
	expect(produk).toBeDefined();

	const refused = await page.request.delete(`/api/produk/${produk!.id}`);
	expect(refused.status()).toBe(409);
	expect((await refused.json()).error).toBe('product_has_sales');

	// …and the catalogue does not offer one either.
	await page.goto('/produk');
	const row = produkRow(page, 'Terjual E2E');
	await expect(row).toContainText('Pernah terjual');
	await expect(row.getByRole('button', { name: 'Hapus' })).toHaveCount(0);
});

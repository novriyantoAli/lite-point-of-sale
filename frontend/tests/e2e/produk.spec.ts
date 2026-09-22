import { expect, test } from '@playwright/test';
import {
	createPengguna,
	createProduk,
	fillProdukForm,
	logIn,
	logOut,
	produkFilter,
	produkForm,
	produkRow
} from './helpers';

/**
 * The catalogue happy paths (ADR-0007, ADR-0009): a real browser, the real
 * SvelteKit BFF, the real Go API on SQLite. Nothing is mocked — including the
 * Kode uniqueness rule, which is enforced by a UNIQUE column and can only be
 * shown to work against a real database.
 *
 * Every test adds its own Produk under a name of its own, because the suite
 * shares one store for the whole run.
 *
 * What is *not* here yet: "Pernah terjual" and the refusal to delete a Produk
 * that sold. Both need a Penjualan to exist, which is #6 — until then the rule
 * is covered by the Go use case tests and by the component test.
 */

test('the Admin adds a Produk and reads it back with its Harga as money', async ({ page }) => {
	await logIn(page);

	await createProduk(page, {
		name: 'Kopi Susu E2E',
		code: 'KOPI-E2E',
		price: 18000,
		stock: 12,
		category: 'Minuman'
	});

	const row = produkRow(page, 'Kopi Susu E2E');
	await expect(row).toContainText('KOPI-E2E');
	await expect(row).toContainText('Minuman');
	await expect(row).toContainText('Rp 18.000');
	await expect(row).toContainText('Aktif');
});

test('a Produk may be added with no Kode and no Kategori at all', async ({ page }) => {
	await logIn(page);

	await createProduk(page, { name: 'Tanpa Kode E2E', price: 3000, stock: 4 });

	// The absent Kode and Kategori are written as an em dash, not left blank or
	// shown as the string "null".
	await expect(produkRow(page, 'Tanpa Kode E2E')).toContainText('—');
});

test('a Kode that is already taken is refused with the API message', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Kode Unik E2E', code: 'UNIK-E2E', price: 1000, stock: 1 });

	await page.getByRole('button', { name: 'Tambah Produk' }).click();
	await fillProdukForm(page, {
		name: 'Kode Duplikat E2E',
		code: 'UNIK-E2E',
		price: 2000,
		stock: 1
	});
	await page.getByRole('button', { name: 'Tambah', exact: true }).click();

	await expect(page.getByRole('alert')).toContainText('Kode sudah dipakai Produk lain.');
	await expect(produkRow(page, 'Kode Duplikat E2E')).toHaveCount(0);
});

test('a Harga left blank is refused before the request is sent', async ({ page }) => {
	await logIn(page);
	await page.goto('/produk');

	await page.getByRole('button', { name: 'Tambah Produk' }).click();
	await fillProdukForm(page, { name: 'Harga Kosong E2E', price: 0, stock: 1 });
	// An empty Harga is a field the Admin forgot, not a Produk priced at 0.
	await produkForm(page).getByLabel('Harga').fill('');
	await page.getByRole('button', { name: 'Tambah', exact: true }).click();

	await expect(page.getByText('Harga harus bilangan bulat.')).toBeVisible();
	await expect(produkRow(page, 'Harga Kosong E2E')).toHaveCount(0);
});

test('the Nama and Kode filters narrow the catalogue', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Filter Kopi E2E', code: 'FILTER-KOPI', price: 1000, stock: 1 });
	await createProduk(page, { name: 'Filter Teh E2E', code: 'FILTER-TEH', price: 2000, stock: 1 });

	const filter = produkFilter(page);

	await filter.getByLabel('Kode').fill('FILTER-TEH');
	await expect(produkRow(page, 'Filter Teh E2E')).toBeVisible();
	await expect(produkRow(page, 'Filter Kopi E2E')).toHaveCount(0);

	await filter.getByLabel('Kode').fill('');
	await filter.getByLabel('Nama').fill('Filter Kopi');
	await expect(produkRow(page, 'Filter Kopi E2E')).toBeVisible();
	await expect(produkRow(page, 'Filter Teh E2E')).toHaveCount(0);
});

test('deactivating a Produk hides it from the Aktif filter and the Nonaktif one finds it', async ({
	page
}) => {
	await logIn(page);
	await createProduk(page, { name: 'Status E2E', price: 5000, stock: 10 });

	await produkRow(page, 'Status E2E').getByRole('button', { name: 'Nonaktifkan' }).click();
	await expect(produkRow(page, 'Status E2E')).toContainText('Nonaktif');

	// The Status dropdown is a bits-ui Select, whose open transition does not run
	// under jsdom — this is the test that drives it for real.
	const filter = produkFilter(page);

	await filter.getByLabel('Status').click();
	await page.getByRole('option', { name: 'Aktif', exact: true }).click();
	await expect(produkRow(page, 'Status E2E')).toHaveCount(0);

	await filter.getByLabel('Status').click();
	await page.getByRole('option', { name: 'Nonaktif', exact: true }).click();
	await expect(produkRow(page, 'Status E2E')).toBeVisible();

	// Back to "Semua": the filter that is not set is not the same as Aktif.
	await filter.getByLabel('Status').click();
	await page.getByRole('option', { name: 'Semua', exact: true }).click();
	await expect(produkRow(page, 'Status E2E')).toBeVisible();
});

test('the Admin changes a Harga from the row of the Produk', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Ubah E2E', price: 18000, stock: 5 });

	await produkRow(page, 'Ubah E2E').getByRole('button', { name: 'Ubah' }).click();

	// The form opens pre-filled with the record being changed.
	const form = produkForm(page);
	await expect(form.getByLabel('Harga')).toHaveValue('18000');
	await form.getByLabel('Harga').fill('22000');
	await page.getByRole('button', { name: 'Simpan Perubahan' }).click();

	await expect(produkRow(page, 'Ubah E2E')).toContainText('Rp 22.000');
});

test('the Admin deletes a Produk that never sold, but only after confirming', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Hapus E2E', price: 1000, stock: 1 });

	await produkRow(page, 'Hapus E2E').getByRole('button', { name: 'Hapus', exact: true }).click();
	// One click asks; nothing is gone yet.
	await expect(produkRow(page, 'Hapus E2E')).toBeVisible();

	await produkRow(page, 'Hapus E2E').getByRole('button', { name: 'Ya, hapus' }).click();

	await expect(page.getByRole('status')).toContainText('Hapus E2E dihapus');
	await expect(produkRow(page, 'Hapus E2E')).toHaveCount(0);
});

test('a Kasir cannot open the catalogue, by link or by URL', async ({ page }) => {
	await logIn(page);
	await createPengguna(page, { username: 'kasir-produk-e2e', password: 'rahasia-kasir' });
	await logOut(page);

	await logIn(page, { username: 'kasir-produk-e2e', password: 'rahasia-kasir' });

	// The entry is hidden…
	await expect(page.getByRole('link', { name: 'Produk' })).toHaveCount(0);
	// …and typing the URL does not get around the guard either.
	await page.goto('/produk');
	await expect(page).toHaveURL('/');
	await expect(page.getByRole('button', { name: 'Tambah Produk' })).toHaveCount(0);

	await logOut(page);
});

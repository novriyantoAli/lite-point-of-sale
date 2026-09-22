import { expect, test } from '@playwright/test';
import {
	createPengguna,
	createProduk,
	logIn,
	logOut,
	stokMenipisSection,
	stokRow,
	addStok
} from './helpers';

/**
 * The Stok happy paths (ADR-0007, ADR-0009): a real browser, the real SvelteKit
 * BFF, the real Go API on SQLite. Nothing is mocked — including the fact that a
 * restock *adds* to the Stok already stored, which is a SQL statement and can
 * only be shown to work against a real database.
 *
 * Every test adds its own Produk under a name of its own, because the suite
 * shares one store for the whole run.
 */

test('the Admin records a restock and the Stok rises by what arrived', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Restok E2E', price: 5000, stock: 0 });

	await addStok(page, 'Restok E2E', 20);

	// The table shows the Stok that is now stored…
	await expect(stokRow(page, 'Restok E2E')).toContainText('20');
	await expect(stokRow(page, 'Restok E2E')).toContainText('Aman');

	// …and the Produk that no longer needs restocking has left the list above it.
	await expect(stokMenipisSection(page).getByText('Restok E2E')).toHaveCount(0);
});

test('two restocks add up instead of the second replacing the first', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Restok Ganda E2E', price: 5000, stock: 1 });

	await addStok(page, 'Restok Ganda E2E', 2);
	await addStok(page, 'Restok Ganda E2E', 3);

	// 1 + 2 + 3, not 3: the form asks for what arrived, never for the new total.
	await expect(stokRow(page, 'Restok Ganda E2E')).toContainText('6');
});

test('a restock of nothing is refused before the request is sent', async ({ page }) => {
	await logIn(page);
	await createProduk(page, { name: 'Restok Kosong E2E', price: 5000, stock: 0 });

	await page.goto('/stok');

	const row = stokRow(page, 'Restok Kosong E2E');
	await row.getByRole('button', { name: 'Tambah Stok' }).click();
	await row.getByLabel('Jumlah masuk').fill('0');
	await row.getByRole('button', { name: 'Tambah Stok' }).click();

	await expect(row.getByText('Jumlah Stok harus lebih dari nol.')).toBeVisible();
	// Refused means refused: the Stok is where it was.
	await expect(row).toContainText('0');
});

test('the restock list holds the Produk that are menipis and states its threshold', async ({
	page
}) => {
	await logIn(page);
	await createProduk(page, { name: 'Menipis E2E', price: 5000, stock: 2 });
	await createProduk(page, { name: 'Aman E2E', price: 5000, stock: 40 });
	// The ambang is the first Stok that is still enough, so a Produk sitting exactly
	// on it is not menipis (issue #5: "di bawah ambang atau nol").
	await createProduk(page, { name: 'Di Ambang E2E', price: 5000, stock: 5 });

	await page.goto('/stok');

	const menipis = stokMenipisSection(page);
	await expect(menipis.getByText('Menipis E2E')).toBeVisible();
	// The rule the list was selected by is the API's to state, so the screen can
	// be read as "these, and why".
	await expect(menipis.getByText(/Stok di bawah 5/)).toBeVisible();
	await expect(menipis.getByText('Aman E2E')).toHaveCount(0);
	await expect(menipis.getByText('Di Ambang E2E')).toHaveCount(0);
});

test('a Kasir cannot open the Stok screen, by link or by URL', async ({ page }) => {
	await logIn(page);
	await createPengguna(page, { username: 'kasir-stok-e2e', password: 'rahasia-kasir' });
	await logOut(page);

	await logIn(page, { username: 'kasir-stok-e2e', password: 'rahasia-kasir' });

	// The entry is hidden…
	await expect(page.getByRole('link', { name: 'Stok' })).toHaveCount(0);
	// …and typing the URL does not get around the guard either: restocking is
	// catalogue management, so it is an Admin screen like /produk.
	await page.goto('/stok');
	await expect(page).toHaveURL('/');
	await expect(page.getByRole('heading', { name: 'Stok' })).toHaveCount(0);

	await logOut(page);
});

import { expect, test } from '@playwright/test';
import {
	bayarTunai,
	bukaKasir,
	createPengguna,
	createProduk,
	logIn,
	logOut,
	nomorStruk,
	printedJob,
	printedSince,
	printerSize,
	tambahProduk
} from './helpers';

/**
 * The Laporan happy paths (#9): a real browser, the real SvelteKit BFF, the real
 * Go API on SQLite. The suite shares one store for the whole run, so the report
 * of *today* already holds every sale the run has made. That is why the omzet
 * assertions compare the screen against what the API answers for the same day
 * rather than against a total this file computes — the aggregation itself is
 * asserted exactly, on a fresh store, in backend/tests/e2e/laporan_test.go.
 *
 * The sales list and the reprint are asserted on this file's own sale, found by
 * its Nomor Struk.
 */

/** Today in the store-local shape the report API reads. */
function hariIni(): string {
	const now = new Date();
	const month = String(now.getMonth() + 1).padStart(2, '0');
	const day = String(now.getDate()).padStart(2, '0');

	return `${now.getFullYear()}-${month}-${day}`;
}

/** An amount as the screen writes it: whole rupiah, grouped in thousands. */
function rupiah(amount: number): string {
	return `Rp ${new Intl.NumberFormat('id-ID').format(amount)}`;
}

test('the Admin reads the omzet and the sales list, and reprints from the list', async ({
	page
}) => {
	await logIn(page);
	await createProduk(page, { name: 'Laporan E2E', price: 9000, stock: 5 });

	await bukaKasir(page);
	await tambahProduk(page, 'Laporan E2E');
	await bayarTunai(page, 9000);
	const nomor = await nomorStruk(page);

	await page.goto('/laporan');

	// What the API aggregated for the day is what the screen shows.
	const answer = await page.request.get(`/api/penjualan/omzet?date=${hariIni()}`);
	expect(answer.status()).toBe(200);
	const report = (await answer.json()).data as { total: number; transactions: number };

	const omzet = page.getByRole('region', { name: 'Omzet harian' });
	await expect(omzet.getByText('Total omzet').locator('..')).toContainText(rupiah(report.total));
	await expect(omzet.getByText('Jumlah transaksi').locator('..')).toContainText(
		String(report.transactions)
	);

	// The breakdown names every method — a zero row included — and the Kasir who
	// rang the sales up.
	for (const metode of ['Tunai', 'QRIS', 'Debit', 'Transfer']) {
		await expect(omzet.getByRole('row', { name: new RegExp(metode) })).toBeVisible();
	}
	await expect(omzet.getByRole('row', { name: /admin/ })).toBeVisible();

	// The sale is in the list, by its Nomor Struk, with the total and the method.
	const daftar = page.getByRole('region', { name: 'Daftar Penjualan' });
	const row = daftar.locator('li').filter({ hasText: `Nomor Struk ${nomor}` });
	await expect(row).toBeVisible();
	await expect(row).toContainText(rupiah(9000));
	await expect(row).toContainText('Tunai');

	// Opening the row reads the same Penjualan the lookup shows, and offers the
	// reprint from there.
	await row.getByRole('button', { name: 'Buka' }).click();
	await expect(row.getByText('Laporan E2E')).toBeVisible();
	await expect(row.locator('p').filter({ hasText: 'Total' })).toContainText('Rp 9.000');

	const before = printerSize();
	await row.getByRole('button', { name: 'Cetak Struk' }).click();
	await expect(row.getByText('Struk tercetak.')).toBeVisible();

	// The bytes of that sale reached the printer file, so the reprint is the real
	// ESC/POS job and not just a success message (ADR-0017, ADR-0009).
	const printed = printedJob(printedSince(before), nomor);
	expect(printed).toContain(`Nomor Struk: ${nomor}`);
	expect(printed).toContain('Laporan E2E');
	expect(printed).toContain('Rp 9.000');
});

test('the Admin narrows the report to a day with no sales', async ({ page }) => {
	await logIn(page);

	await page.goto('/laporan');
	await page.getByLabel('Tanggal').fill('2000-01-01');

	// An empty day is answered, not hidden: nothing was sold, and the four methods
	// are still there as zero rows.
	const omzet = page.getByRole('region', { name: 'Omzet harian' });
	await expect(omzet.getByText('Belum ada Penjualan pada tanggal ini.')).toBeVisible();
	await expect(omzet.getByText('Total omzet').locator('..')).toContainText('Rp 0');

	const daftar = page.getByRole('region', { name: 'Daftar Penjualan' });
	await expect(daftar.getByText('Belum ada Penjualan pada tanggal ini.')).toBeVisible();
});

test('a Kasir has no Laporan entry, and is turned away from the screen and the API', async ({
	page
}) => {
	await logIn(page);
	await createPengguna(page, { username: 'kasir-laporan-e2e', password: 'rahasia-kasir' });
	await logOut(page);

	await logIn(page, { username: 'kasir-laporan-e2e', password: 'rahasia-kasir' });

	// Revenue is not the till's screen: the menu does not offer it…
	await expect(page.getByRole('link', { name: 'Laporan' })).toHaveCount(0);

	// …the page guard turns the URL away, and the API refuses it too.
	await page.goto('/laporan');
	await expect(page).toHaveURL('/');

	const refused = await page.request.get(`/api/penjualan/omzet?date=${hariIni()}`);
	expect(refused.status()).toBe(403);
	expect((await refused.json()).error).toBe('forbidden');
});

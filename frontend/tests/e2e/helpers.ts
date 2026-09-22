import { expect, type Page } from '@playwright/test';

/**
 * The credentials of the seeded Admin Pengguna, matching the
 * POS_ADMIN_USERNAME/POS_ADMIN_PASSWORD the Go API is started with in
 * playwright.config.ts. The store is wiped by global-setup.ts first, so the
 * seed runs on every E2E run with exactly these credentials.
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

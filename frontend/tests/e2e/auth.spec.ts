import { expect, test } from '@playwright/test';
import { ADMIN, createPengguna, logIn, logOut, penggunaRow } from './helpers';

/**
 * The authentication happy paths (ADR-0007, ADR-0009): a real browser, the real
 * SvelteKit BFF, the real Go API on SQLite. Nothing is mocked — including the
 * bcrypt hashing and the session cookie, which are the parts this slice exists
 * for.
 */

test('an anonymous visitor is sent to the login page', async ({ page }) => {
	await page.goto('/');

	await expect(page).toHaveURL('/login');
	await expect(page.getByRole('button', { name: 'Masuk' })).toBeVisible();
});

test('a wrong password keeps the terminal on the login form', async ({ page }) => {
	await page.goto('/login');

	await page.getByLabel('Username').fill(ADMIN.username);
	await page.getByLabel('Password').fill('salah-sekali');
	await page.getByRole('button', { name: 'Masuk' }).click();

	await expect(page.getByRole('alert')).toHaveText('Username atau password salah.');
	await expect(page).toHaveURL('/login');
});

test('the Admin creates a Kasir who can log in but cannot open the Pengguna page', async ({
	page
}) => {
	await logIn(page);
	await createPengguna(page, { username: 'kasir-e2e', password: 'rahasia-kasir' });
	await logOut(page);

	await logIn(page, { username: 'kasir-e2e', password: 'rahasia-kasir' });

	// The Admin entry is hidden…
	await expect(page.getByRole('link', { name: 'Pengguna' })).toHaveCount(0);
	// …and typing the URL does not get around it either.
	await page.goto('/pengguna');
	await expect(page).toHaveURL('/');
	await expect(page.getByText('Pengguna', { exact: true })).toHaveCount(0);

	await logOut(page);
});

test('an Admin who deactivates a Kasir ends that Kasir login for good', async ({ page }) => {
	await logIn(page);
	await createPengguna(page, {
		username: 'kasir-nonaktif',
		password: 'rahasia-kasir',
		role: 'Admin'
	});

	// The Admin's own row offers no way to lock themselves out.
	await expect(penggunaRow(page, ADMIN.username)).toContainText(
		'Tidak bisa menonaktifkan diri sendiri'
	);

	await penggunaRow(page, 'kasir-nonaktif').getByRole('button', { name: 'Nonaktifkan' }).click();
	await expect(penggunaRow(page, 'kasir-nonaktif')).toContainText('Nonaktif');

	await logOut(page);

	await page.goto('/login');
	await page.getByLabel('Username').fill('kasir-nonaktif');
	await page.getByLabel('Password').fill('rahasia-kasir');
	await page.getByRole('button', { name: 'Masuk' }).click();

	await expect(page.getByRole('alert')).toHaveText('Username atau password salah.');
	await expect(page).toHaveURL('/login');
});

test('logging out ends the session for the next caller of the terminal', async ({ page }) => {
	await logIn(page);
	await logOut(page);

	// The cookie is gone, so even the dashboard is out of reach again.
	await page.goto('/');
	await expect(page).toHaveURL('/login');
});

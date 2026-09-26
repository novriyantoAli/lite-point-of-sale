import { render, screen, within } from '@testing-library/svelte';
import { createRawSnippet } from 'svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import Layout from './+layout.svelte';

const { storeName, logout, goto, currentUrl } = vi.hoisted(() => ({
	storeName: vi.fn(),
	logout: vi.fn(),
	goto: vi.fn(),
	currentUrl: { pathname: '/kasir' }
}));

vi.mock('$app/state', () => ({ page: { url: currentUrl } }));
vi.mock('$app/navigation', () => ({ goto }));
/**
 * The group prefix is what `resolve` strips, and it is what the layout compares
 * against `page.url.pathname` to decide the active tab — so the mock has to
 * strip it too, or every tab would read as inactive.
 */
vi.mock('$app/paths', () => ({ resolve: (path: string) => path.replace('/(app)', '') }));
// The barrel is the seam: the rail reads the store's name through the pengaturan
// domain's public surface and never through its `api/` (ADR-0006).
vi.mock('$lib/domains/pengaturan', () => ({ createStoreNameQuery: storeName }));
vi.mock('$lib/domains/auth/api/auth.api', () => ({ authApi: { logout } }));

const children = createRawSnippet(() => ({ render: () => '<p>isi layar</p>' }));

const kasir = { id: 2, username: 'kasir1', role: 'kasir' as const, active: true };
const admin = { id: 1, username: 'admin', role: 'admin' as const, active: true };

// The file is named without the `+` SvelteKit reserves for routes, so it tests
// `./+layout.svelte` — the rail every screen in this group sits inside.
describe('(app) layout — rel navigasi', () => {
	beforeEach(() => {
		storeName.mockReset();
		logout.mockReset();
		goto.mockReset();
		currentUrl.pathname = '/kasir';
	});

	it('leads with the store name and marks the tab that is open', () => {
		storeName.mockReturnValue({ isPending: false, data: 'Toko Elektronik Jaya' });

		render(Layout, { data: { user: kasir }, children }, { wrapper: QueryClientHarness });

		// The store's own name is the rail's identity, not the product's name
		// (PRODUCT.md, dikonfirmasi pemilik).
		const rail = screen.getByRole('banner');
		expect(within(rail).getByText('Toko Elektronik Jaya')).toBeInTheDocument();
		expect(within(rail).queryByText('Lite Point of Sale')).not.toBeInTheDocument();

		// `/kasir` is the open screen, and only that tab is filled with the red.
		expect(screen.getByRole('link', { name: 'Kasir' })).toHaveAttribute('aria-current', 'page');
		expect(screen.getByRole('link', { name: 'Beranda' })).not.toHaveAttribute('aria-current');
	});

	it('falls back to the product name while the store has none', () => {
		storeName.mockReturnValue({ isPending: false, data: '' });

		render(Layout, { data: { user: kasir }, children }, { wrapper: QueryClientHarness });

		expect(within(screen.getByRole('banner')).getByText('Lite Point of Sale')).toBeInTheDocument();
	});

	it('shows nothing rather than the wrong name while the name loads', () => {
		storeName.mockReturnValue({ isPending: true, data: undefined });

		render(Layout, { data: { user: kasir }, children }, { wrapper: QueryClientHarness });

		expect(screen.queryByText('Lite Point of Sale')).not.toBeInTheDocument();
		expect(screen.getByText('Memuat nama toko…')).toBeInTheDocument();
	});

	it('offers a Kasir only the three screens a Kasir may open', () => {
		storeName.mockReturnValue({ isPending: false, data: 'Toko Elektronik Jaya' });

		render(Layout, { data: { user: kasir }, children }, { wrapper: QueryClientHarness });

		expect(screen.getByRole('link', { name: 'Beranda' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Penjualan' })).toBeInTheDocument();
		expect(screen.queryByRole('link', { name: 'Produk' })).not.toBeInTheDocument();
		expect(screen.queryByRole('link', { name: 'Backup' })).not.toBeInTheDocument();
	});

	it('adds the Admin row for an Admin', () => {
		storeName.mockReturnValue({ isPending: false, data: 'Toko Elektronik Jaya' });

		render(Layout, { data: { user: admin }, children }, { wrapper: QueryClientHarness });

		expect(screen.getByRole('link', { name: 'Produk' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Backup' })).toBeInTheDocument();
	});

	it('shows who is logged in and how to leave, in the rail', () => {
		storeName.mockReturnValue({ isPending: false, data: 'Toko Elektronik Jaya' });

		render(Layout, { data: { user: kasir }, children }, { wrapper: QueryClientHarness });

		// Who is logged in is the rail's business; which Peran they hold is the
		// SessionMenu's own test's business.
		const rail = screen.getByRole('banner');
		expect(within(rail).getByText('kasir1')).toBeInTheDocument();
		expect(within(rail).getByRole('button', { name: 'Keluar' })).toBeInTheDocument();
	});

	it('renders the screen it wraps', () => {
		storeName.mockReturnValue({ isPending: false, data: 'Toko Elektronik Jaya' });

		render(Layout, { data: { user: kasir }, children }, { wrapper: QueryClientHarness });

		expect(screen.getByText('isi layar')).toBeInTheDocument();
	});
});

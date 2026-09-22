import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import SessionMenu from './SessionMenu.svelte';

const { logout, goto } = vi.hoisted(() => ({ logout: vi.fn(), goto: vi.fn() }));

vi.mock('../api/auth.api', () => ({ authApi: { logout } }));
vi.mock('$app/navigation', () => ({ goto }));
vi.mock('$app/paths', () => ({ resolve: (path: string) => path }));

const kasir = { id: 2, username: 'kasir1', role: 'kasir' as const, active: true };

describe('SessionMenu', () => {
	beforeEach(() => {
		logout.mockReset();
		goto.mockReset();
	});

	it('shows who is logged in and with which Peran', () => {
		render(SessionMenu, { user: kasir }, { wrapper: QueryClientHarness });

		expect(screen.getByText('kasir1')).toBeInTheDocument();
		expect(screen.getByText('Kasir')).toBeInTheDocument();
	});

	it('logs out and hands the terminal back to the login page', async () => {
		logout.mockResolvedValue(undefined);
		const user = userEvent.setup();

		render(SessionMenu, { user: kasir }, { wrapper: QueryClientHarness });

		await user.click(screen.getByRole('button', { name: 'Keluar' }));

		expect(logout).toHaveBeenCalledOnce();
		expect(goto).toHaveBeenCalledWith('/login', { invalidateAll: true });
	});

	it('still leaves the terminal on the login page when the call fails', async () => {
		logout.mockRejectedValue({ message: 'Tidak dapat menghubungi server.', status: 502 });
		const user = userEvent.setup();

		render(SessionMenu, { user: kasir }, { wrapper: QueryClientHarness });

		await user.click(screen.getByRole('button', { name: 'Keluar' }));

		expect(goto).toHaveBeenCalledWith('/login', { invalidateAll: true });
	});
});

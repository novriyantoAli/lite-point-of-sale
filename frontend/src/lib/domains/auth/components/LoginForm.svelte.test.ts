import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import LoginForm from './LoginForm.svelte';

const { login, goto } = vi.hoisted(() => ({ login: vi.fn(), goto: vi.fn() }));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/auth.api', () => ({ authApi: { login } }));
vi.mock('$app/navigation', () => ({ goto }));
vi.mock('$app/paths', () => ({ resolve: (path: string) => path }));

const pengguna = { id: 1, username: 'admin', role: 'admin', active: true };

describe('LoginForm', () => {
	beforeEach(() => {
		login.mockReset();
		goto.mockReset();
	});

	it('logs in and sends the terminal to the dashboard', async () => {
		login.mockResolvedValue(pengguna);
		const user = userEvent.setup();

		render(LoginForm, {}, { wrapper: QueryClientHarness });

		await user.type(screen.getByLabelText('Username'), 'admin');
		await user.type(screen.getByLabelText('Password'), 'rahasia123');
		await user.click(screen.getByRole('button', { name: 'Masuk' }));

		expect(login).toHaveBeenCalledWith({ username: 'admin', password: 'rahasia123' });
		expect(goto).toHaveBeenCalledWith('/', { invalidateAll: true });
	});

	it('shows the message the API sent and stays on the form', async () => {
		login.mockRejectedValue({
			message: 'Username atau password salah.',
			status: 401,
			code: 'invalid_credentials'
		});
		const user = userEvent.setup();

		render(LoginForm, {}, { wrapper: QueryClientHarness });

		await user.type(screen.getByLabelText('Username'), 'admin');
		await user.type(screen.getByLabelText('Password'), 'salah');
		await user.click(screen.getByRole('button', { name: 'Masuk' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Username atau password salah.');
		expect(goto).not.toHaveBeenCalled();
	});

	it('refuses an empty form without asking the API', async () => {
		const user = userEvent.setup();

		render(LoginForm, {}, { wrapper: QueryClientHarness });

		await user.click(screen.getByRole('button', { name: 'Masuk' }));

		expect(await screen.findByText('Username wajib diisi.')).toBeInTheDocument();
		expect(screen.getByText('Password wajib diisi.')).toBeInTheDocument();
		expect(login).not.toHaveBeenCalled();
	});
});

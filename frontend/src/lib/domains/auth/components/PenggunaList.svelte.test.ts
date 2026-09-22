import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import PenggunaList from './PenggunaList.svelte';

const { listPengguna, createPengguna, setPenggunaActive } = vi.hoisted(() => ({
	listPengguna: vi.fn(),
	createPengguna: vi.fn(),
	setPenggunaActive: vi.fn()
}));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/auth.api', () => ({
	authApi: { listPengguna, createPengguna, setPenggunaActive }
}));

const admin = { id: 1, username: 'admin', role: 'admin' as const, active: true };
const kasir = { id: 2, username: 'kasir1', role: 'kasir' as const, active: true };

function renderList(currentUserId = admin.id) {
	return render(PenggunaList, { currentUserId }, { wrapper: QueryClientHarness });
}

describe('PenggunaList', () => {
	beforeEach(() => {
		listPengguna.mockReset();
		createPengguna.mockReset();
		setPenggunaActive.mockReset();
	});

	it('lists every Pengguna with its Peran', async () => {
		listPengguna.mockResolvedValue([admin, kasir]);

		renderList();

		expect(await screen.findByText('kasir1')).toBeInTheDocument();
		expect(screen.getByText('Admin')).toBeInTheDocument();
		expect(screen.getByText('Kasir')).toBeInTheDocument();
	});

	it('offers no way to deactivate the Admin using the page', async () => {
		listPengguna.mockResolvedValue([admin, kasir]);

		renderList();

		expect(await screen.findByText('(Anda)')).toBeInTheDocument();
		expect(screen.getByText('Tidak bisa menonaktifkan diri sendiri')).toBeInTheDocument();
		// One button only: the row of the other Pengguna.
		expect(screen.getAllByRole('button', { name: 'Nonaktifkan' })).toHaveLength(1);
	});

	it('deactivates another Pengguna and marks them Nonaktif', async () => {
		listPengguna.mockResolvedValue([admin, kasir]);
		setPenggunaActive.mockResolvedValue({ ...kasir, active: false });
		const user = userEvent.setup();

		renderList();

		await user.click(await screen.findByRole('button', { name: 'Nonaktifkan' }));

		expect(setPenggunaActive).toHaveBeenCalledWith(2, false);
	});

	it('offers to reactivate a Nonaktif Pengguna', async () => {
		listPengguna.mockResolvedValue([admin, { ...kasir, active: false }]);
		const user = userEvent.setup();

		renderList();

		expect(await screen.findByText('Nonaktif')).toBeInTheDocument();

		await user.click(screen.getByRole('button', { name: 'Aktifkan' }));

		expect(setPenggunaActive).toHaveBeenCalledWith(2, true);
	});

	it('shows the normalized error and a retry when the list fails', async () => {
		listPengguna.mockRejectedValue({
			message: 'Anda tidak berhak melakukan tindakan ini.',
			status: 403
		});
		const user = userEvent.setup();

		renderList();

		expect(
			await screen.findByText('Anda tidak berhak melakukan tindakan ini.')
		).toBeInTheDocument();

		await user.click(screen.getByRole('button', { name: 'Coba lagi' }));

		expect(listPengguna).toHaveBeenCalledTimes(2);
	});

	it('says so when the list comes back empty, instead of showing nothing', async () => {
		listPengguna.mockResolvedValue([]);

		renderList();

		expect(await screen.findByText(/Belum ada Pengguna lain/)).toBeInTheDocument();
	});

	it('creates a Pengguna from the form and reports it', async () => {
		listPengguna.mockResolvedValue([admin]);
		createPengguna.mockResolvedValue(kasir);
		const user = userEvent.setup();

		renderList();

		await user.type(screen.getByLabelText('Username'), 'kasir1');
		await user.type(screen.getByLabelText('Password'), 'rahasia123');
		await user.click(screen.getByRole('button', { name: 'Tambah' }));

		expect(createPengguna).toHaveBeenCalledWith({
			username: 'kasir1',
			password: 'rahasia123',
			role: 'kasir'
		});
		expect(await screen.findByRole('status')).toHaveTextContent('Pengguna kasir1 ditambahkan.');
	});

	it('refuses a password below the minimum without asking the API', async () => {
		listPengguna.mockResolvedValue([admin]);
		const user = userEvent.setup();

		renderList();

		await user.type(screen.getByLabelText('Username'), 'kasir1');
		await user.type(screen.getByLabelText('Password'), 'pendek');
		await user.click(screen.getByRole('button', { name: 'Tambah' }));

		expect(await screen.findByText('Password minimal 8 karakter.')).toBeInTheDocument();
		expect(createPengguna).not.toHaveBeenCalled();
	});

	it('shows the message the API sent when creation is refused', async () => {
		listPengguna.mockResolvedValue([admin]);
		createPengguna.mockRejectedValue({ message: 'Username sudah dipakai.', status: 409 });
		const user = userEvent.setup();

		renderList();

		await user.type(screen.getByLabelText('Username'), 'kasir1');
		await user.type(screen.getByLabelText('Password'), 'rahasia123');
		await user.click(screen.getByRole('button', { name: 'Tambah' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Username sudah dipakai.');
	});
});

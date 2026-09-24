import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import PengaturanForm from './PengaturanForm.svelte';

const { get, update } = vi.hoisted(() => ({
	get: vi.fn(),
	update: vi.fn()
}));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/pengaturan.api', () => ({
	pengaturanApi: { get, update }
}));

const pengaturan = {
	id: 1,
	header: 'Toko Kopi\nJl. Melati 1',
	footer: 'Terima kasih',
	paper_width: 80,
	low_stock_threshold: 5
};

function renderForm() {
	return render(PengaturanForm, {}, { wrapper: QueryClientHarness });
}

describe('PengaturanForm', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('says it is loading before the Pengaturan arrives', () => {
		get.mockReturnValue(new Promise(() => {}));

		renderForm();

		expect(screen.getByText('Memuat Pengaturan…')).toBeInTheDocument();
	});

	it('shows the normalized error and a retry that works', async () => {
		get.mockRejectedValueOnce({ message: 'Tidak dapat menghubungi server.', status: 502 });
		const user = userEvent.setup();

		renderForm();

		expect(await screen.findByText('Tidak dapat menghubungi server.')).toBeInTheDocument();

		get.mockResolvedValue(pengaturan);
		await user.click(screen.getByRole('button', { name: 'Coba lagi' }));

		expect(await screen.findByLabelText('Header Struk')).toHaveValue('Toko Kopi\nJl. Melati 1');
	});

	it('seeds the form from the stored Pengaturan', async () => {
		get.mockResolvedValue(pengaturan);

		renderForm();

		expect(await screen.findByLabelText('Header Struk')).toHaveValue('Toko Kopi\nJl. Melati 1');
		expect(screen.getByLabelText('Footer Struk')).toHaveValue('Terima kasih');
		expect(screen.getByLabelText('Ambang Stok menipis')).toHaveValue('5');
	});

	it('saves the changed Pengaturan and reports it', async () => {
		get.mockResolvedValue(pengaturan);
		update.mockResolvedValue({ ...pengaturan, low_stock_threshold: 3 });
		const user = userEvent.setup();

		renderForm();

		await screen.findByLabelText('Ambang Stok menipis');
		await user.clear(screen.getByLabelText('Ambang Stok menipis'));
		await user.type(screen.getByLabelText('Ambang Stok menipis'), '3');
		await user.click(screen.getByRole('button', { name: 'Simpan Pengaturan' }));

		expect(update).toHaveBeenCalledWith({
			header: 'Toko Kopi\nJl. Melati 1',
			footer: 'Terima kasih',
			paper_width: 80,
			low_stock_threshold: 3
		});
		expect(await screen.findByRole('status')).toHaveTextContent('Pengaturan disimpan.');
	});

	it('refuses a blank ambang with a message, without asking the API', async () => {
		get.mockResolvedValue(pengaturan);
		const user = userEvent.setup();

		renderForm();

		await screen.findByLabelText('Ambang Stok menipis');
		await user.clear(screen.getByLabelText('Ambang Stok menipis'));
		await user.click(screen.getByRole('button', { name: 'Simpan Pengaturan' }));

		expect(
			await screen.findByText('Ambang Stok menipis harus bilangan bulat.')
		).toBeInTheDocument();
		expect(update).not.toHaveBeenCalled();
	});

	it('shows the message the API sent when it refuses the input', async () => {
		get.mockResolvedValue(pengaturan);
		update.mockRejectedValue({ message: 'Lebar kertas harus 58 atau 80 mm.', status: 400 });
		const user = userEvent.setup();

		renderForm();

		await screen.findByLabelText('Ambang Stok menipis');
		await user.click(screen.getByRole('button', { name: 'Simpan Pengaturan' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Lebar kertas harus 58 atau 80 mm.');
	});
});

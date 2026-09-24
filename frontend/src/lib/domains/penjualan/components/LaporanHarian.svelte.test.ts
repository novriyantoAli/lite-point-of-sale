import { render, screen, waitFor, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import LaporanHarian from './LaporanHarian.svelte';
import { laporanState } from '../state/laporan.state.svelte';

const { list, omzetHarian, getByReceiptNumber, cetakStruk } = vi.hoisted(() => ({
	list: vi.fn(),
	omzetHarian: vi.fn(),
	getByReceiptNumber: vi.fn(),
	cetakStruk: vi.fn()
}));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/penjualan.api', () => ({
	penjualanApi: { list, omzetHarian, getByReceiptNumber, cetakStruk }
}));

const omzet = {
	date: '2026-09-23',
	total: 54000,
	transactions: 2,
	by_method: [
		{ method: 'cash', total: 18000, transactions: 1 },
		{ method: 'qris', total: 36000, transactions: 1 },
		{ method: 'debit', total: 0, transactions: 0 },
		{ method: 'transfer', total: 0, transactions: 0 }
	],
	by_cashier: [{ cashier_id: 1, cashier_name: 'kasir1', total: 54000, transactions: 2 }]
};

const rows = [
	{
		receipt_number: 2,
		created_at: '2026-09-23 11:00:00',
		cashier_id: 1,
		cashier_name: 'kasir1',
		total: 36000,
		method: 'qris'
	},
	{
		receipt_number: 1,
		created_at: '2026-09-23 10:00:00',
		cashier_id: 1,
		cashier_name: 'kasir1',
		total: 18000,
		method: 'cash'
	}
];

const sale = {
	receipt_number: 2,
	created_at: '2026-09-23 11:00:00',
	cashier_id: 1,
	cashier_name: 'kasir1',
	total: 36000,
	items: [{ product_id: 1, name: 'Kopi Susu', price: 18000, quantity: 2, subtotal: 36000 }],
	payment: { method: 'qris', amount: 36000, change: 0 }
};

function omzetSection() {
	return within(screen.getByRole('region', { name: 'Omzet harian' }));
}

function daftarSection() {
	return within(screen.getByRole('region', { name: 'Daftar Penjualan' }));
}

function renderLaporan() {
	return render(LaporanHarian, {}, { wrapper: QueryClientHarness });
}

describe('LaporanHarian', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		laporanState.reset();
		list.mockResolvedValue(rows);
		omzetHarian.mockResolvedValue(omzet);
		getByReceiptNumber.mockResolvedValue(sale);
		cetakStruk.mockResolvedValue({ printed: true });
	});

	it('shows the omzet of the day: total, transactions and both breakdowns', async () => {
		renderLaporan();

		// The total card and the per-Kasir row can carry the same amount, so the
		// figure is read off the card it belongs to.
		const totalCard = (await omzetSection().findByText('Total omzet')).closest('div');
		expect(within(totalCard!).getByText('Rp 54.000')).toBeInTheDocument();

		// Every method the API answered, in the till's order — the ones nobody used
		// read as zero rather than going missing.
		for (const label of ['Tunai', 'QRIS', 'Debit', 'Transfer']) {
			expect(omzetSection().getByText(label)).toBeInTheDocument();
		}

		expect(omzetSection().getByText('kasir1')).toBeInTheDocument();
	});

	it('lists the Penjualan of the day with their Nomor Struk', async () => {
		renderLaporan();

		expect(await daftarSection().findByText('2')).toBeInTheDocument();
		expect(daftarSection().getByText('1')).toBeInTheDocument();
		// The row names the Kasir, the method and the time it was rung up.
		expect(daftarSection().getAllByText(/Kasir kasir1/)).toHaveLength(2);
		expect(daftarSection().getAllByText(/QRIS/)).toHaveLength(1);
	});

	it('says it is loading before either answer arrives', () => {
		list.mockReturnValue(new Promise(() => {}));
		omzetHarian.mockReturnValue(new Promise(() => {}));

		renderLaporan();

		expect(screen.getByText('Memuat omzet…')).toBeInTheDocument();
		expect(screen.getByText('Memuat Penjualan…')).toBeInTheDocument();
	});

	it('shows the normalized error and a retry that works', async () => {
		omzetHarian.mockRejectedValueOnce({
			message: 'Anda tidak berhak melakukan tindakan ini.',
			status: 403
		});
		const user = userEvent.setup();

		renderLaporan();

		expect(
			await omzetSection().findByText('Anda tidak berhak melakukan tindakan ini.')
		).toBeInTheDocument();

		omzetHarian.mockResolvedValue(omzet);
		await user.click(omzetSection().getByRole('button', { name: 'Coba lagi' }));

		const totalCard = (await omzetSection().findByText('Total omzet')).closest('div');
		expect(within(totalCard!).getByText('Rp 54.000')).toBeInTheDocument();
	});

	it('answers an empty day without pretending there were sales', async () => {
		list.mockResolvedValue([]);
		omzetHarian.mockResolvedValue({
			date: '2026-09-23',
			total: 0,
			transactions: 0,
			by_method: [
				{ method: 'cash', total: 0, transactions: 0 },
				{ method: 'qris', total: 0, transactions: 0 },
				{ method: 'debit', total: 0, transactions: 0 },
				{ method: 'transfer', total: 0, transactions: 0 }
			],
			by_cashier: []
		});

		renderLaporan();

		expect(
			await daftarSection().findByText('Belum ada Penjualan pada tanggal ini.')
		).toBeInTheDocument();
		expect(omzetSection().getByText('Belum ada Penjualan pada tanggal ini.')).toBeInTheDocument();
		expect(
			omzetSection().getByText('Belum ada Penjualan, jadi belum ada yang bisa diatribusikan.')
		).toBeInTheDocument();
		// The method breakdown is still complete: four zero rows.
		for (const label of ['Tunai', 'QRIS', 'Debit', 'Transfer']) {
			expect(omzetSection().getByText(label)).toBeInTheDocument();
		}
	});

	it('opens a Penjualan from the list and reprints its Struk', async () => {
		const user = userEvent.setup();

		renderLaporan();

		const row = (await daftarSection().findByText('2')).closest('li');
		expect(row).not.toBeNull();

		await user.click(within(row!).getByRole('button', { name: 'Buka' }));

		// The record the lookup shows, read by the same Nomor Struk.
		expect(await within(row!).findByText('Kopi Susu')).toBeInTheDocument();
		await waitFor(() => expect(getByReceiptNumber).toHaveBeenCalledWith(2));

		await user.click(within(row!).getByRole('button', { name: 'Cetak Struk' }));

		await waitFor(() => expect(cetakStruk).toHaveBeenCalledWith(2));
		expect(await within(row!).findByText('Struk tercetak.')).toBeInTheDocument();
	});
});

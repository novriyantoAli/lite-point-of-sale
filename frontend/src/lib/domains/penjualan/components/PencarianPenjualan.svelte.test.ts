import { render, screen, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import PencarianPenjualan from './PencarianPenjualan.svelte';

const { getByReceiptNumber } = vi.hoisted(() => ({ getByReceiptNumber: vi.fn() }));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/penjualan.api', () => ({
	penjualanApi: { checkout: vi.fn(), getByReceiptNumber }
}));

/** A Penjualan stored under Nomor Struk 7, paid in Tunai with 14000 back. */
const sale = {
	receipt_number: 7,
	created_at: '2026-09-23 10:00:00',
	cashier_id: 1,
	cashier_name: 'kasir1',
	total: 36000,
	items: [{ product_id: 1, name: 'Kopi Susu', price: 18000, quantity: 2, subtotal: 36000 }],
	payment: { method: 'cash', amount: 50000, change: 14000 }
};

function renderScreen() {
	return render(PencarianPenjualan, {}, { wrapper: QueryClientHarness });
}

/** Types a Nomor Struk and submits the lookup the way a person does. */
async function cari(user: ReturnType<typeof userEvent.setup>, nomor: string) {
	await user.type(screen.getByLabelText('Nomor Struk'), nomor);
	await user.click(screen.getByRole('button', { name: 'Cari' }));
}

/** The found record, scoped so "Nomor Struk" in the form is never matched. */
async function hasil() {
	return within(await screen.findByRole('region', { name: 'Penjualan tersimpan' }));
}

beforeEach(() => {
	vi.clearAllMocks();
});

describe('PencarianPenjualan', () => {
	it('says what to do before anything has been searched for, and asks nothing', () => {
		renderScreen();

		expect(screen.getByText(/Ketik Nomor Struk lalu tekan Cari/)).toBeInTheDocument();
		expect(getByReceiptNumber).not.toHaveBeenCalled();
	});

	it('shows a skeleton, not a blank screen, while Go is answering', async () => {
		getByReceiptNumber.mockReturnValue(new Promise(() => {}));
		const user = userEvent.setup();

		renderScreen();
		await cari(user, '7');

		expect(await screen.findByRole('status')).toHaveTextContent('Mencari Penjualan…');
	});

	it('shows the stored Penjualan: items, total, Pembayaran, and the Kembalian for Tunai', async () => {
		getByReceiptNumber.mockResolvedValue(sale);
		const user = userEvent.setup();

		renderScreen();
		await cari(user, '7');

		expect(getByReceiptNumber).toHaveBeenCalledWith(7);

		const record = await hasil();

		// The Nomor Struk, when, and who rang it up.
		expect(record.getByText(/Nomor Struk/)).toHaveTextContent('Nomor Struk 7');
		expect(record.getByText(/Kasir kasir1/)).toBeInTheDocument();

		// What was sold, at the price it was sold for (CONTEXT.md, Item).
		expect(record.getByText(/Kopi Susu/)).toBeInTheDocument();
		expect(record.getByText(/2 × Rp 18\.000/)).toBeInTheDocument();
		expect(record.getByText('Total').parentElement).toHaveTextContent('Rp 36.000');

		// How it was paid, and the Kembalian only Tunai produces (CONTEXT.md,
		// Kembalian).
		expect(record.getByText('Bayar · Tunai').parentElement).toHaveTextContent('Rp 50.000');
		expect(record.getByText('Kembalian').parentElement).toHaveTextContent('Rp 14.000');
	});

	it('shows no Kembalian for a recorded method', async () => {
		const qris = { ...sale, payment: { method: 'qris', amount: 36000, change: 0 } };
		getByReceiptNumber.mockResolvedValue(qris);
		const user = userEvent.setup();

		renderScreen();
		await cari(user, '7');

		const record = await hasil();

		expect(record.getByText('Bayar · QRIS').parentElement).toHaveTextContent('Rp 36.000');
		expect(record.queryByText('Kembalian')).not.toBeInTheDocument();
	});

	it('reads a different Nomor Struk when a second search is submitted', async () => {
		const other = {
			...sale,
			receipt_number: 8,
			items: [{ ...sale.items[0]!, name: 'Teh Botol', subtotal: 12000 }],
			total: 12000,
			payment: { method: 'qris', amount: 12000, change: 0 }
		};
		getByReceiptNumber.mockImplementation((nomorStruk: number) =>
			Promise.resolve(nomorStruk === 8 ? other : sale)
		);
		const user = userEvent.setup();

		renderScreen();
		await cari(user, '7');
		await screen.findByText(/Kopi Susu/);

		// The second lookup reuses the same mounted record, so the query has to
		// follow the new number rather than keep the first answer on screen.
		await user.clear(screen.getByLabelText('Nomor Struk'));
		await cari(user, '8');

		expect(await screen.findByText(/Teh Botol/)).toBeInTheDocument();
		expect(screen.queryByText(/Kopi Susu/)).not.toBeInTheDocument();
	});

	it('reads a Nomor Struk that names nothing as a message, not an empty screen', async () => {
		getByReceiptNumber.mockRejectedValueOnce({
			message: 'Penjualan tidak ditemukan.',
			status: 404,
			code: 'sale_not_found'
		});
		const user = userEvent.setup();

		renderScreen();
		await cari(user, '999');

		expect(await screen.findByRole('alert')).toHaveTextContent('Penjualan tidak ditemukan.');
		expect(screen.queryByRole('region', { name: 'Penjualan tersimpan' })).not.toBeInTheDocument();
	});

	it('retries the same Nomor Struk from the error, and then shows the Penjualan', async () => {
		getByReceiptNumber.mockRejectedValueOnce({
			message: 'Tidak dapat menghubungi server.',
			status: 502
		});
		getByReceiptNumber.mockResolvedValue(sale);
		const user = userEvent.setup();

		renderScreen();
		await cari(user, '7');
		await screen.findByRole('alert');

		await user.click(screen.getByRole('button', { name: 'Coba lagi' }));

		const record = await hasil();
		expect(record.getByText(/Kopi Susu/)).toBeInTheDocument();
	});

	it('refuses a field that is not a number, without asking the API', async () => {
		const user = userEvent.setup();

		renderScreen();
		await cari(user, 'abc');

		expect(await screen.findByText('Nomor Struk harus bilangan bulat.')).toBeInTheDocument();
		expect(getByReceiptNumber).not.toHaveBeenCalled();
	});

	it('refuses zero: a Nomor Struk starts at one', async () => {
		const user = userEvent.setup();

		renderScreen();
		await cari(user, '0');

		expect(await screen.findByText('Nomor Struk harus lebih dari nol.')).toBeInTheDocument();
		expect(getByReceiptNumber).not.toHaveBeenCalled();
	});
});

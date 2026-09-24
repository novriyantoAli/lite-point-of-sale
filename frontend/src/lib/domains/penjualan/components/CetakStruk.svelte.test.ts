import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import CetakStruk from './CetakStruk.svelte';

const { cetakStruk } = vi.hoisted(() => ({ cetakStruk: vi.fn() }));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/penjualan.api', () => ({
	penjualanApi: { checkout: vi.fn(), getByReceiptNumber: vi.fn(), cetakStruk }
}));

function renderCetak(hasilAwal: { printed: boolean; message?: string } | null = null) {
	return render(CetakStruk, { nomorStruk: 7, hasilAwal }, { wrapper: QueryClientHarness });
}

beforeEach(() => {
	vi.clearAllMocks();
});

describe('CetakStruk', () => {
	it('offers to print, and reports nothing before it has', () => {
		renderCetak();

		expect(screen.getByRole('button', { name: 'Cetak Struk' })).toBeEnabled();
		expect(screen.queryByRole('status')).not.toBeInTheDocument();
		expect(screen.queryByRole('alert')).not.toBeInTheDocument();
		expect(cetakStruk).not.toHaveBeenCalled();
	});

	it('prints the Struk of the Nomor Struk it was given, and says so', async () => {
		cetakStruk.mockResolvedValue({ printed: true });
		const user = userEvent.setup();

		renderCetak();
		await user.click(screen.getByRole('button', { name: 'Cetak Struk' }));

		expect(cetakStruk).toHaveBeenCalledWith(7);
		expect(await screen.findByRole('status')).toHaveTextContent('Struk tercetak.');
		expect(screen.queryByRole('alert')).not.toBeInTheDocument();
	});

	it('shows the outcome of the automatic print it was given, before anything is pressed', () => {
		renderCetak({ printed: false, message: 'Printer belum diatur.' });

		// The Kasir sees that the Struk did not come out without having to press
		// anything, and the button offers the retry.
		expect(screen.getByRole('alert')).toHaveTextContent('Printer belum diatur.');
		expect(screen.getByRole('button', { name: 'Cetak ulang Struk' })).toBeEnabled();
		expect(cetakStruk).not.toHaveBeenCalled();
	});

	it('clears the failure once the retry prints', async () => {
		cetakStruk.mockResolvedValue({ printed: true });
		const user = userEvent.setup();

		renderCetak({ printed: false, message: 'Printer belum diatur.' });
		await user.click(screen.getByRole('button', { name: 'Cetak ulang Struk' }));

		expect(cetakStruk).toHaveBeenCalledWith(7);
		expect(await screen.findByRole('status')).toHaveTextContent('Struk tercetak.');
		expect(screen.queryByRole('alert')).not.toBeInTheDocument();
	});

	it('keeps a failed retry on screen with the button still offered', async () => {
		cetakStruk.mockResolvedValue({ printed: false, message: 'Kertas habis.' });
		const user = userEvent.setup();

		renderCetak({ printed: false, message: 'Printer belum diatur.' });
		await user.click(screen.getByRole('button', { name: 'Cetak ulang Struk' }));

		// The alert follows the mutation's own answer, so it takes a tick to replace
		// the automatic print's message.
		await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Kertas habis.'));
		// The button comes back once the retry has settled, offering another go.
		expect(await screen.findByRole('button', { name: 'Cetak ulang Struk' })).toBeEnabled();
	});

	it('shows one alert when the retry request itself fails after a failed print', async () => {
		cetakStruk.mockRejectedValue({ message: 'Tidak dapat menghubungi server.', status: 502 });
		const user = userEvent.setup();

		renderCetak({ printed: false, message: 'Printer belum diatur.' });
		await user.click(screen.getByRole('button', { name: 'Cetak ulang Struk' }));

		// The request that could not be made is the latest news, and it does not stack
		// on top of the print result it never replaced.
		await waitFor(() =>
			expect(screen.getByRole('alert')).toHaveTextContent('Tidak dapat menghubungi server.')
		);
		expect(screen.getAllByRole('alert')).toHaveLength(1);
	});

	it('shows the API error when the reprint itself could not be made', async () => {
		cetakStruk.mockRejectedValue({
			message: 'Penjualan tidak ditemukan.',
			status: 404,
			code: 'sale_not_found'
		});
		const user = userEvent.setup();

		renderCetak();
		await user.click(screen.getByRole('button', { name: 'Cetak Struk' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Penjualan tidak ditemukan.');
		await waitFor(() => expect(cetakStruk).toHaveBeenCalledTimes(1));
	});
});

import { render, screen, waitFor, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import StokList from './StokList.svelte';

const { list, lowStock, addStock } = vi.hoisted(() => ({
	list: vi.fn(),
	lowStock: vi.fn(),
	addStock: vi.fn()
}));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/produk.api', () => ({
	produkApi: { list, lowStock, addStock }
}));

const teh = {
	id: 2,
	name: 'Teh Botol',
	code: null,
	price: 5000,
	category: null,
	stock: 0,
	active: true,
	sold: false
};

const kopi = {
	id: 1,
	name: 'Kopi Susu',
	code: 'KOPI-01',
	price: 18000,
	category: 'Minuman',
	stock: 2,
	active: true,
	sold: false
};

const air = {
	id: 3,
	name: 'Air Mineral',
	code: null,
	price: 3000,
	category: null,
	stock: 40,
	active: true,
	sold: false
};

/** A Produk that is not for sale: its Stok cannot run out in a way that matters. */
const roti = {
	id: 4,
	name: 'Roti Bakar',
	code: null,
	price: 15000,
	category: null,
	stock: 0,
	active: false,
	sold: false
};

/** A Produk sitting exactly on the threshold: the first Stok that is still enough. */
const diAmbang = {
	id: 5,
	name: 'Di Ambang',
	code: null,
	price: 4000,
	category: null,
	stock: 5,
	active: true,
	sold: false
};

/**
 * The two lists carry the same Produk on purpose, so a test has to say which one
 * it is reading — otherwise "Tambah Stok" would match a button in each.
 */
function restockList() {
	return within(screen.getByRole('region', { name: 'Stok menipis' }));
}

function catalogue() {
	return within(screen.getByRole('region', { name: 'Stok per Produk' }));
}

/** One row of the per-Produk table, so a badge is read off the Produk it belongs to. */
function rowOf(name: string) {
	return within(catalogue().getByRole('row', { name: new RegExp(name) }));
}

function renderStok() {
	return render(StokList, {}, { wrapper: QueryClientHarness });
}

describe('StokList', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		list.mockResolvedValue([kopi, teh, air, roti, diAmbang]);
		lowStock.mockResolvedValue({ threshold: 5, products: [teh, kopi] });
	});

	it('lists what to restock, with the threshold the API selected by', async () => {
		renderStok();

		expect(await restockList().findByText('Teh Botol')).toBeInTheDocument();
		expect(restockList().getByText('Kopi Susu')).toBeInTheDocument();
		// The rule the list was selected by is the API's to state: a threshold the
		// UI invented could disagree with the list underneath it.
		expect(restockList().getByText(/Stok di bawah 5/)).toBeInTheDocument();
	});

	it('says it is loading before either list arrives', () => {
		list.mockReturnValue(new Promise(() => {}));
		lowStock.mockReturnValue(new Promise(() => {}));

		renderStok();

		expect(screen.getByText('Memuat Stok menipis…')).toBeInTheDocument();
		expect(screen.getByText('Memuat Produk…')).toBeInTheDocument();
	});

	it('shows the normalized error and a retry that works', async () => {
		lowStock.mockRejectedValueOnce({
			message: 'Anda tidak berhak melakukan tindakan ini.',
			status: 403
		});
		const user = userEvent.setup();

		renderStok();

		expect(
			await restockList().findByText('Anda tidak berhak melakukan tindakan ini.')
		).toBeInTheDocument();

		lowStock.mockResolvedValue({ threshold: 5, products: [teh, kopi] });
		await user.click(restockList().getByRole('button', { name: 'Coba lagi' }));

		expect(await restockList().findByText('Teh Botol')).toBeInTheDocument();
	});

	it('says there is nothing to restock when the list is empty', async () => {
		lowStock.mockResolvedValue({ threshold: 5, products: [] });

		renderStok();

		expect(
			await restockList().findByText(/Tidak ada Produk dengan Stok menipis/)
		).toBeInTheDocument();
	});

	it('says the catalogue is empty when there is nothing to show Stok for', async () => {
		list.mockResolvedValue([]);

		renderStok();

		expect(await catalogue().findByText(/Belum ada Produk/)).toBeInTheDocument();
	});

	it('tells habis, menipis, aman and Nonaktif apart', async () => {
		renderStok();

		await catalogue().findByText('Teh Botol');

		expect(rowOf('Teh Botol').getByText('Habis')).toBeInTheDocument();
		expect(rowOf('Kopi Susu').getByText('Menipis')).toBeInTheDocument();
		expect(rowOf('Air Mineral').getByText('Aman')).toBeInTheDocument();
		// The ambang is the first Stok that is still enough, so the Produk sitting
		// exactly on it is aman, not menipis (issue #5: "di bawah ambang atau nol").
		expect(rowOf('Di Ambang').getByText('Aman')).toBeInTheDocument();
		// A Produk that is not for sale is not called habis here: it does not run
		// out, it is simply not sold (CONTEXT.md, Nonaktif).
		expect(rowOf('Roti Bakar').getByText('Nonaktif')).toBeInTheDocument();
		// The restock list says the same thing in the words of the person reading it.
		expect(restockList().getByText('Stok habis')).toBeInTheDocument();
	});

	it('does not call a Produk aman while the threshold it needs is unknown', async () => {
		lowStock.mockRejectedValue({ message: 'Tidak dapat menghubungi server.', status: 502 });

		renderStok();

		await catalogue().findByText('Kopi Susu');

		// Kopi Susu has Stok 2, which is menipis — but the rule that decides that is
		// the one that could not be read, so the table says nothing rather than
		// reassuring the Admin about a Stok it cannot judge.
		expect(rowOf('Kopi Susu').queryByText('Aman')).toBeNull();
		expect(rowOf('Kopi Susu').getByText('—')).toBeInTheDocument();

		// Stok 0 is habis by arithmetic, not by the threshold, so it still reads.
		expect(rowOf('Teh Botol').getByText('Habis')).toBeInTheDocument();
	});

	it('restocks the Produk whose row asked for it', async () => {
		lowStock.mockResolvedValue({ threshold: 5, products: [teh] });
		addStock.mockResolvedValue({ ...teh, stock: 6 });
		const user = userEvent.setup();

		renderStok();

		await user.click(await restockList().findByRole('button', { name: 'Tambah Stok' }));
		await user.type(restockList().getByLabelText('Jumlah masuk'), '6');
		await user.click(restockList().getByRole('button', { name: 'Tambah Stok' }));

		expect(addStock).toHaveBeenCalledWith(2, 6);
		expect(await screen.findByRole('status')).toHaveTextContent('Stok Teh Botol sekarang 6.');
	});

	it('opens the restock form only in the list it was asked from', async () => {
		lowStock.mockResolvedValue({ threshold: 5, products: [teh] });
		const user = userEvent.setup();

		renderStok();

		await user.click(await restockList().findByRole('button', { name: 'Tambah Stok' }));

		// The same Produk sits in both lists; one delivery means one form, or the
		// Admin types into one and the other keeps a stale amount.
		expect(restockList().getByLabelText('Jumlah masuk')).toBeInTheDocument();
		expect(catalogue().queryByLabelText('Jumlah masuk')).toBeNull();

		// Every row keeps its own button: the form moved into one list, not away from
		// the other.
		const row = catalogue().getByRole('row', { name: /Teh Botol/ });
		expect(within(row).getByRole('button', { name: 'Tambah Stok' })).toBeInTheDocument();
	});

	it('refuses a blank quantity with a message, without asking the API', async () => {
		lowStock.mockResolvedValue({ threshold: 5, products: [teh] });
		const user = userEvent.setup();

		renderStok();

		await user.click(await restockList().findByRole('button', { name: 'Tambah Stok' }));
		await user.click(restockList().getByRole('button', { name: 'Tambah Stok' }));

		expect(await restockList().findByText('Jumlah Stok harus bilangan bulat.')).toBeInTheDocument();
		expect(addStock).not.toHaveBeenCalled();
	});

	it('refuses a quantity of zero: a delivery that did not happen is not a restock', async () => {
		lowStock.mockResolvedValue({ threshold: 5, products: [teh] });
		const user = userEvent.setup();

		renderStok();

		await user.click(await restockList().findByRole('button', { name: 'Tambah Stok' }));
		await user.type(restockList().getByLabelText('Jumlah masuk'), '0');
		await user.click(restockList().getByRole('button', { name: 'Tambah Stok' }));

		expect(await restockList().findByText('Jumlah Stok harus lebih dari nol.')).toBeInTheDocument();
		expect(addStock).not.toHaveBeenCalled();
	});

	it('shows the message the API sent when the restock is refused', async () => {
		lowStock.mockResolvedValue({ threshold: 5, products: [teh] });
		addStock.mockRejectedValue({ message: 'Produk tidak ditemukan.', status: 404 });
		const user = userEvent.setup();

		renderStok();

		await user.click(await restockList().findByRole('button', { name: 'Tambah Stok' }));
		await user.type(restockList().getByLabelText('Jumlah masuk'), '3');
		await user.click(restockList().getByRole('button', { name: 'Tambah Stok' }));

		expect(await restockList().findByRole('alert')).toHaveTextContent('Produk tidak ditemukan.');
	});

	it('leaves the list showing the Stok the API added to', async () => {
		lowStock.mockResolvedValue({ threshold: 5, products: [teh] });
		addStock.mockResolvedValue({ ...teh, stock: 20 });
		const user = userEvent.setup();

		renderStok();

		await user.click(await restockList().findByRole('button', { name: 'Tambah Stok' }));
		await user.type(restockList().getByLabelText('Jumlah masuk'), '20');

		// The restock list is refetched, so the Produk that is no longer menipis
		// leaves it — the screen must not keep showing an alert the API dropped.
		lowStock.mockResolvedValue({ threshold: 5, products: [] });
		await user.click(restockList().getByRole('button', { name: 'Tambah Stok' }));

		await waitFor(() =>
			expect(restockList().getByText(/Tidak ada Produk dengan Stok menipis/)).toBeInTheDocument()
		);
	});
});

import { render, screen, waitFor, within } from '@testing-library/svelte';
import userEvent, { type UserEvent } from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import ProdukList from './ProdukList.svelte';
import { produkFilterState } from '../state/produk.state.svelte';

const { list, categories, create, update, setActive, remove, lowStock } = vi.hoisted(() => ({
	list: vi.fn(),
	categories: vi.fn(),
	create: vi.fn(),
	update: vi.fn(),
	setActive: vi.fn(),
	remove: vi.fn(),
	lowStock: vi.fn()
}));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/produk.api', () => ({
	produkApi: { list, categories, create, update, setActive, remove, lowStock }
}));

const kopi = {
	id: 1,
	name: 'Kopi Susu',
	code: 'KOPI-01',
	price: 18000,
	category: 'Minuman',
	stock: 12,
	active: true,
	sold: false
};

/** A Produk that already sold: it may be deactivated, never deleted. */
const air = {
	id: 2,
	name: 'Air Mineral',
	code: null,
	price: 3000,
	category: null,
	stock: 40,
	active: false,
	sold: true
};

/** A Produk the domain's own ambang puts below the line. */
const teh = {
	id: 3,
	name: 'Teh Manis',
	code: 'TEH-01',
	price: 5000,
	category: null,
	stock: 3,
	active: true,
	sold: false
};

/**
 * The filter bar and the Produk form both have a field labelled "Nama", and both
 * a "Kode" and a "Kategori" — deliberately, because that is what each is to the
 * person using it. Scoping by the labelled region is what keeps a test honest
 * about which one it typed into.
 */
function filterBar() {
	return within(screen.getByRole('search', { name: 'Saring katalog' }));
}

function form() {
	return within(screen.getByRole('form', { name: 'Formulir Produk' }));
}

function renderCatalogue() {
	return render(ProdukList, {}, { wrapper: QueryClientHarness });
}

/**
 * Opens the Produk form — adding from the strip, or changing from a row — and
 * waits for the dialog to have taken focus.
 *
 * The dialog is a modal, so it traps focus when it opens, and that trap lands a
 * tick after the click that opened it. A keystroke issued in between is
 * delivered to the dialog instead of to the field. Only a test can be that
 * fast, which is why the wait lives here rather than in the component.
 */
async function openForm(user: UserEvent, trigger: 'Tambah Produk' | 'Ubah' = 'Tambah Produk') {
	await user.click(await screen.findByRole('button', { name: trigger }));

	await waitFor(() =>
		expect(document.querySelector('[data-slot="dialog-content"]')).toContainElement(
			document.activeElement as HTMLElement
		)
	);
}

describe('ProdukList', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		// The filter state is a module singleton, so it outlives a test unless it
		// is put back — exactly like the QueryClient the harness rebuilds.
		produkFilterState.reset();
		categories.mockResolvedValue([]);
		// Nothing is menipis unless a test says so: the ambang is the domain's answer,
		// and the catalogue only ever writes the word it is handed.
		lowStock.mockResolvedValue({ threshold: 5, products: [] });
	});

	it('lists the catalogue, writing Harga as money rather than a bare number', async () => {
		list.mockResolvedValue([kopi]);

		renderCatalogue();

		expect(await screen.findByText('Kopi Susu')).toBeInTheDocument();
		expect(screen.getByText('Rp 18.000')).toBeInTheDocument();
		expect(screen.getByText('KOPI-01')).toBeInTheDocument();
		expect(screen.getByText('Minuman')).toBeInTheDocument();
	});

	it('says it is loading before the catalogue arrives', () => {
		list.mockReturnValue(new Promise(() => {}));

		renderCatalogue();

		expect(screen.getByText('Memuat katalog…')).toBeInTheDocument();
	});

	it('shows the normalized error and a retry that works', async () => {
		list.mockRejectedValueOnce({
			message: 'Anda tidak berhak melakukan tindakan ini.',
			status: 403
		});
		const user = userEvent.setup();

		renderCatalogue();

		expect(
			await screen.findByText('Anda tidak berhak melakukan tindakan ini.')
		).toBeInTheDocument();

		list.mockResolvedValue([kopi]);
		await user.click(screen.getByRole('button', { name: 'Coba lagi' }));

		expect(await screen.findByText('Kopi Susu')).toBeInTheDocument();
	});

	it('says the catalogue is empty when nothing is filtered', async () => {
		list.mockResolvedValue([]);

		renderCatalogue();

		expect(await screen.findByText(/Belum ada Produk/)).toBeInTheDocument();
	});

	it('tells a filter that matched nothing apart from an empty catalogue', async () => {
		list.mockResolvedValue([]);
		produkFilterState.name = 'tidak ada';

		renderCatalogue();

		expect(await screen.findByText(/Tidak ada Produk yang cocok/)).toBeInTheDocument();
	});

	it('writes the word for menipis from the ambang the domain gave it, and none other', async () => {
		list.mockResolvedValue([kopi, teh]);
		lowStock.mockResolvedValue({ threshold: 5, products: [teh] });

		renderCatalogue();

		// Teh is below the ambang the domain reported, so it carries the word beside
		// its number — the same word the Stok screen's list uses.
		expect(await screen.findByText('3 · menipis')).toBeInTheDocument();
		// Kopi is not on that list, so it is just its number: the catalogue does not
		// decide menipis for itself and would disagree with the Stok screen if it did.
		expect(screen.getByText('12')).toBeInTheDocument();
	});

	it('writes a blank Kode and Kategori as words, not as an absent cell', async () => {
		list.mockResolvedValue([air]);

		renderCatalogue();

		// A blank Kode column must not read as data that went missing, and "null" is
		// never a word an Admin should be shown (DESIGN.md, Do's).
		expect(await screen.findByText('tanpa Kode')).toBeInTheDocument();
		expect(screen.getByText('—')).toBeInTheDocument();

		// Status carries its word too: the quiet tag's trade dress is not the whole
		// message, and it is the same vocabulary the kasir tiles use.
		expect(screen.getByText('Nonaktif')).toBeInTheDocument();
	});

	it('adds a Produk from the form and reports it', async () => {
		list.mockResolvedValue([]);
		create.mockResolvedValue(kopi);
		const user = userEvent.setup();

		renderCatalogue();

		await openForm(user);
		await user.type(form().getByLabelText('Nama'), 'Kopi Susu');
		await user.type(form().getByLabelText('Harga'), '18000');
		await user.type(form().getByLabelText('Stok'), '12');
		await user.click(screen.getByRole('button', { name: 'Tambah' }));

		// A blank Kode and Kategori are normalized to null, the value the API stores,
		// and a new Produk carries the Status it starts with.
		expect(create).toHaveBeenCalledWith({
			name: 'Kopi Susu',
			code: null,
			price: 18000,
			category: null,
			stock: 12,
			active: true
		});
		expect(await screen.findByRole('status')).toHaveTextContent('Produk Kopi Susu disimpan.');
	});

	it('refuses a blank Harga with a message, without asking the API', async () => {
		list.mockResolvedValue([]);
		const user = userEvent.setup();

		renderCatalogue();

		await openForm(user);
		await user.type(form().getByLabelText('Nama'), 'Kopi');
		await user.type(form().getByLabelText('Stok'), '1');
		await user.click(screen.getByRole('button', { name: 'Tambah' }));

		expect(await screen.findByText('Harga harus bilangan bulat.')).toBeInTheDocument();
		expect(create).not.toHaveBeenCalled();
	});

	it('shows the message the API sent when the Kode is already taken', async () => {
		list.mockResolvedValue([]);
		create.mockRejectedValue({ message: 'Kode sudah dipakai Produk lain.', status: 409 });
		const user = userEvent.setup();

		renderCatalogue();

		await openForm(user);
		await user.type(form().getByLabelText('Nama'), 'Kopi Susu');
		await user.type(form().getByLabelText('Harga'), '18000');
		await user.type(form().getByLabelText('Stok'), '1');
		await user.click(screen.getByRole('button', { name: 'Tambah' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Kode sudah dipakai Produk lain.');
	});

	it('opens the row of a Produk pre-filled and saves the change', async () => {
		list.mockResolvedValue([kopi]);
		update.mockResolvedValue({ ...kopi, price: 22000 });
		const user = userEvent.setup();

		renderCatalogue();

		await openForm(user, 'Ubah');

		const harga = form().getByLabelText('Harga');
		expect(harga).toHaveValue('18000');

		// Neither Stok nor Status is an edit-form field: Stok moves through the Stok
		// awal of a create, the Tambah Stok form, and a Penjualan (ADR-0014), and
		// Status through the Aktifkan/Nonaktifkan button.
		expect(form().queryByLabelText('Stok')).toBeNull();
		expect(form().queryByLabelText('Status')).toBeNull();

		await user.clear(harga);
		await user.type(harga, '22000');
		await user.click(screen.getByRole('button', { name: 'Simpan Perubahan' }));

		// The edit body carries the editable record only — no Stok to write back over a
		// delivery that arrived while the form was open.
		expect(update).toHaveBeenCalledWith(1, {
			name: 'Kopi Susu',
			code: 'KOPI-01',
			price: 22000,
			category: 'Minuman'
		});
	});

	it('deactivates a Produk without deleting it', async () => {
		list.mockResolvedValue([kopi]);
		setActive.mockResolvedValue({ ...kopi, active: false });
		const user = userEvent.setup();

		renderCatalogue();

		await user.click(await screen.findByRole('button', { name: 'Nonaktifkan' }));

		expect(setActive).toHaveBeenCalledWith(1, false);
		expect(remove).not.toHaveBeenCalled();
	});

	it('offers no delete for a Produk that already sold', async () => {
		list.mockResolvedValue([air]);

		renderCatalogue();

		expect(await screen.findByText('Pernah terjual')).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'Hapus' })).not.toBeInTheDocument();
	});

	it('deletes only after a second, deliberate click', async () => {
		list.mockResolvedValue([kopi]);
		remove.mockResolvedValue(undefined);
		const user = userEvent.setup();

		renderCatalogue();

		await user.click(await screen.findByRole('button', { name: 'Hapus' }));
		expect(remove).not.toHaveBeenCalled();

		await user.click(screen.getByRole('button', { name: 'Ya, hapus' }));

		expect(remove).toHaveBeenCalledWith(1);
		expect(await screen.findByRole('status')).toHaveTextContent('Produk Kopi Susu dihapus.');
	});

	it('shows the refusal when the API will not delete a Produk that sold', async () => {
		list.mockResolvedValue([kopi]);
		remove.mockRejectedValue({
			message: 'Produk yang sudah pernah terjual hanya bisa dinonaktifkan.',
			status: 409
		});
		const user = userEvent.setup();

		renderCatalogue();

		await user.click(await screen.findByRole('button', { name: 'Hapus' }));
		await user.click(screen.getByRole('button', { name: 'Ya, hapus' }));

		expect(await screen.findByRole('alert')).toHaveTextContent(
			'Produk yang sudah pernah terjual hanya bisa dinonaktifkan.'
		);
	});

	it('asks for the catalogue the Nama filter describes', async () => {
		list.mockResolvedValue([kopi]);
		const user = userEvent.setup();

		renderCatalogue();

		await screen.findByText('Kopi Susu');
		await user.type(filterBar().getByLabelText('Nama'), 'kopi');

		await waitFor(() =>
			expect(list).toHaveBeenLastCalledWith(expect.objectContaining({ name: 'kopi' }))
		);
	});

	it('shows the filters as words, not as the raw values behind them', async () => {
		list.mockResolvedValue([kopi]);

		renderCatalogue();

		await screen.findByText('Kopi Susu');

		// The Kategori Select holds the sentinel `semua` when nothing is picked;
		// what an Admin has to read is the label.
		expect(filterBar().getByLabelText('Kategori')).toHaveTextContent('Semua Kategori');
		expect(filterBar().getByLabelText('Status')).toHaveTextContent(/^Semua$/);
	});

	it('asks again when the Status filter changes, keeping Nonaktif apart from Semua', async () => {
		list.mockResolvedValue([kopi]);

		renderCatalogue();

		await screen.findByText('Kopi Susu');

		// The Select itself is bits-ui's and its open transition does not run under
		// jsdom — the e2e suite is what drives the dropdown. What is worth pinning
		// here is the state it writes: `false` has to reach the API as a filter
		// rather than be dropped as falsy, which is what "Semua" would be.
		produkFilterState.active = false;

		await waitFor(() =>
			expect(list).toHaveBeenLastCalledWith(expect.objectContaining({ active: false }))
		);
	});
});

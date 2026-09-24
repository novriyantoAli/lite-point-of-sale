import MockAdapter from 'axios-mock-adapter';
import { render, screen, waitFor, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { apiClient } from '$lib/api/client';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import { keranjangState } from '../state/keranjang.state.svelte';
import Kasir from './Kasir.svelte';

/**
 * The api/client seam is what is mocked here, not a domain's `api/` module: the
 * till reads the produk domain through its barrel, and reaching past the barrel
 * to fake it is exactly what ADR-0006 forbids. Mocking the client keeps the real
 * api layers — and their schema parsing — in the test (ADR-0007).
 */
let mock: MockAdapter;

const kopi = {
	id: 1,
	name: 'Kopi Susu',
	code: 'KOPI-01',
	price: 18000,
	category: null,
	stock: 10,
	active: true,
	sold: false
};

const teh = {
	id: 2,
	name: 'Teh Botol',
	code: null,
	price: 6000,
	category: null,
	stock: 2,
	active: true,
	sold: false
};

/** A Produk that cannot be sold yet: there is nothing to take out of its Stok. */
const habis = {
	id: 3,
	name: 'Air Mineral',
	code: null,
	price: 3000,
	category: null,
	stock: 0,
	active: true,
	sold: false
};

const catalogue = [kopi, teh, habis];

/** The Penjualan the API answers a checkout of 2 × Kopi Susu with, paid 50000. */
const sale = {
	receipt_number: 1,
	created_at: '2026-09-23 10:00:00',
	cashier_id: 1,
	cashier_name: 'kasir1',
	total: 36000,
	items: [{ product_id: 1, name: 'Kopi Susu', price: 18000, quantity: 2, subtotal: 36000 }],
	payment: { method: 'cash', amount: 50000, change: 14000 }
};

type LookupParams = { name?: string; code?: string; active?: string };

/**
 * Serves the till lookup, narrowing the way `GET /api/produk?active=true` does, so
 * a test that types a Kode sees the listing it would really see.
 */
function serveCatalogue(produk = catalogue) {
	mock.onGet('/produk').reply((config) => {
		const params = (config.params ?? {}) as LookupParams;

		return [
			200,
			{
				data: produk.filter(
					(candidate) =>
						(!params.code || (candidate.code ?? '').includes(params.code)) &&
						(!params.name || candidate.name.toLowerCase().includes(params.name.toLowerCase()))
				)
			}
		];
	});
}

function keranjang() {
	return within(screen.getByRole('region', { name: 'Keranjang' }));
}

function pembayaran() {
	return within(screen.getByRole('form', { name: 'Pembayaran' }));
}

function struk() {
	return within(screen.getByRole('region', { name: 'Penjualan tercatat' }));
}

function renderKasir() {
	return render(Kasir, {}, { wrapper: QueryClientHarness });
}

/** Adds a Produk to the keranjang the way the Kasir does: the Tambah button. */
async function tambah(user: ReturnType<typeof userEvent.setup>, name: string) {
	await user.click(screen.getByRole('button', { name: `Tambah ${name} ke keranjang` }));
}

beforeEach(() => {
	mock = new MockAdapter(apiClient);
	keranjangState.clear();
});

afterEach(() => {
	mock.restore();
});

describe('Kasir', () => {
	it('offers the Produk that are for sale, asking only for the Aktif ones', async () => {
		serveCatalogue();

		renderKasir();

		expect(await screen.findByText(/Kopi Susu/)).toBeInTheDocument();
		expect(screen.getByText('Teh Botol')).toBeInTheDocument();
		// `active: true` is the whole of the kasir lookup: a Nonaktif Produk is not
		// for sale, so it is never offered (CONTEXT.md, Nonaktif).
		expect(mock.history.get[0]!.params).toEqual({ active: 'true' });
	});

	it('says there is nothing to sell when the catalogue has no Aktif Produk', async () => {
		serveCatalogue([]);

		renderKasir();

		expect(await screen.findByText(/Belum ada Produk yang bisa dijual/)).toBeInTheDocument();
	});

	it('shows the normalized error and a retry that works', async () => {
		mock.onGet('/produk').replyOnce(403, {
			message: 'Anda tidak berhak melakukan tindakan ini.',
			error: 'forbidden'
		});
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();

		expect(
			await screen.findByText('Anda tidak berhak melakukan tindakan ini.')
		).toBeInTheDocument();

		await user.click(screen.getByRole('button', { name: 'Coba lagi' }));

		expect(await screen.findByText(/Kopi Susu/)).toBeInTheDocument();
	});

	it('adds a Produk to the keranjang and keeps the running total', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);

		await tambah(user, 'Kopi Susu');

		expect(keranjang().getByText('1 unit dalam 1 Item.')).toBeInTheDocument();
		expect(keranjang().getByRole('status')).toHaveTextContent('Rp 18.000');

		// One more unit: the total follows the quantity (CONTEXT.md, Item).
		await user.click(keranjang().getByRole('button', { name: 'Tambah jumlah Kopi Susu' }));

		expect(keranjang().getByText('2 unit dalam 1 Item.')).toBeInTheDocument();
		expect(keranjang().getByRole('status')).toHaveTextContent('Rp 36.000');
	});

	it('keeps one line per Produk, however many times it is added', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);

		await tambah(user, 'Kopi Susu');
		await tambah(user, 'Kopi Susu');

		// One Produk is one Item (CONTEXT.md, Item) — never two lines of the same
		// Produk, which the API would have to merge anyway.
		expect(keranjang().getAllByRole('listitem')).toHaveLength(1);
		expect(keranjang().getByLabelText('Jumlah Kopi Susu')).toHaveValue(2);
	});

	it('sets a quantity by typing it, and writes back the quantity it kept', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');

		const jumlah = keranjang().getByLabelText('Jumlah Kopi Susu');
		await user.clear(jumlah);
		await user.type(jumlah, '3');

		expect(jumlah).toHaveValue(3);
		expect(keranjang().getByRole('status')).toHaveTextContent('Rp 54.000');

		// A quantity the keranjang cannot hold is not stored, and the field is put
		// back to what is: the subtotal beside it must not disagree with it.
		await user.clear(jumlah);
		await user.type(jumlah, '0');

		expect(jumlah).toHaveValue(3);
		expect(keranjang().getByRole('status')).toHaveTextContent('Rp 54.000');
	});

	it('removes a line, and empties the whole keranjang', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');
		await tambah(user, 'Teh Botol');

		await user.click(keranjang().getByRole('button', { name: 'Hapus Kopi Susu dari keranjang' }));

		expect(keranjang().queryByLabelText('Jumlah Kopi Susu')).not.toBeInTheDocument();
		expect(keranjang().getByLabelText('Jumlah Teh Botol')).toBeInTheDocument();

		await user.click(keranjang().getByRole('button', { name: 'Kosongkan' }));

		expect(keranjang().getByText(/Keranjang kosong/)).toBeInTheDocument();
	});

	it('blocks the checkout while a line is above the Stok', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText('Teh Botol');
		await tambah(user, 'Teh Botol');

		// Teh has 2 in Stok; three units is the refusal of CONTEXT.md, Stok.
		await user.click(keranjang().getByRole('button', { name: 'Tambah jumlah Teh Botol' }));
		await user.click(keranjang().getByRole('button', { name: 'Tambah jumlah Teh Botol' }));

		expect(keranjang().getByText(/Melebihi Stok: tersisa 2/)).toBeInTheDocument();
		expect(
			pembayaran().getByText('Ada Item yang melebihi Stok. Kurangi jumlahnya lebih dulu.')
		).toBeInTheDocument();
		expect(pembayaran().getByRole('button', { name: 'Bayar & Simpan Penjualan' })).toBeDisabled();
	});

	it('does not offer a Produk whose Stok is habis', async () => {
		serveCatalogue();

		renderKasir();

		await screen.findByText('Air Mineral');

		expect(screen.getByRole('button', { name: 'Tambah Air Mineral ke keranjang' })).toBeDisabled();
	});

	it('shows the Kembalian as the amount is typed', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');
		await tambah(user, 'Kopi Susu');

		await user.type(pembayaran().getByLabelText('Jumlah bayar'), '50000');

		// Kembalian = bayar − total (CONTEXT.md, Kembalian).
		expect(pembayaran().getByRole('status')).toHaveTextContent('Rp 14.000');
	});

	it('refuses an amount below the total before the request is sent', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');

		await user.type(pembayaran().getByLabelText('Jumlah bayar'), '1000');
		await user.click(pembayaran().getByRole('button', { name: 'Bayar & Simpan Penjualan' }));

		expect(await pembayaran().findByText('Jumlah bayar kurang dari total.')).toBeInTheDocument();
		expect(mock.history.post).toHaveLength(0);
	});

	it('refuses a blank amount rather than reading it as 0', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');

		await user.click(pembayaran().getByRole('button', { name: 'Bayar & Simpan Penjualan' }));

		expect(await pembayaran().findByText('Jumlah bayar harus bilangan bulat.')).toBeInTheDocument();
		expect(mock.history.post).toHaveLength(0);
	});

	it('offers the four Pembayaran methods, with Tunai the one already picked', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');

		// Tunai is the default: it is the method that needs an amount typed, and the
		// one a till takes most of the day.
		expect(pembayaran().getByRole('radio', { name: 'Tunai' })).toBeChecked();
		for (const label of ['QRIS', 'Debit', 'Transfer']) {
			expect(pembayaran().getByRole('radio', { name: label })).not.toBeChecked();
		}
	});

	it('records a non-tunai Pembayaran for the total, with nothing to type and no Kembalian', async () => {
		serveCatalogue();
		const qris = { ...sale, payment: { method: 'qris', amount: 36000, change: 0 } };
		mock.onPost('/penjualan').reply(201, { data: { sale: qris } });
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');
		await tambah(user, 'Kopi Susu');

		await user.click(pembayaran().getByRole('radio', { name: 'QRIS' }));

		// A recorded method pays the total, so there is no amount for the Kasir to
		// type — and no Kembalian for the API to work out (CONTEXT.md, Kembalian).
		expect(pembayaran().queryByLabelText('Jumlah bayar')).not.toBeInTheDocument();
		expect(pembayaran().queryByText('Kembalian')).not.toBeInTheDocument();
		expect(pembayaran().getByRole('status')).toHaveTextContent('QRIS · Rp 36.000');

		await user.click(pembayaran().getByRole('button', { name: 'Bayar & Simpan Penjualan' }));

		await waitFor(() => expect(mock.history.post).toHaveLength(1));
		expect(JSON.parse(mock.history.post[0]!.data)).toEqual({
			items: [{ product_id: 1, quantity: 2 }],
			payment: { method: 'qris', amount: 36000 }
		});

		// The Struk is the API's answer, and a recorded method has no Kembalian to
		// show.
		expect(await screen.findByText(/Nomor Struk/)).toBeInTheDocument();
		expect(struk().getByText(/Bayar · QRIS/)).toBeInTheDocument();
		expect(struk().queryByText('Kembalian')).not.toBeInTheDocument();
	});

	it('checks the keranjang out and shows the Penjualan that was recorded', async () => {
		serveCatalogue();
		mock.onPost('/penjualan').reply(201, { data: { sale } });
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');
		await tambah(user, 'Kopi Susu');

		await user.type(pembayaran().getByLabelText('Jumlah bayar'), '50000');
		await user.click(pembayaran().getByRole('button', { name: 'Bayar & Simpan Penjualan' }));

		// The cart goes out as the Produk id and the quantity, and the method as
		// Tunai — nothing the catalogue is the one to know, and no Kembalian the
		// client would be declaring.
		await waitFor(() => expect(mock.history.post).toHaveLength(1));
		expect(JSON.parse(mock.history.post[0]!.data)).toEqual({
			items: [{ product_id: 1, quantity: 2 }],
			payment: { method: 'cash', amount: 50000 }
		});

		// What is shown is the API's answer: the Nomor Struk and the Kembalian are
		// its numbers, not this screen's arithmetic.
		expect(await screen.findByText(/Nomor Struk/)).toHaveTextContent('Nomor Struk 1');
		expect(struk().getByText(/Kopi Susu/)).toBeInTheDocument();
		expect(struk().getByText('Rp 14.000')).toBeInTheDocument();

		// The keranjang is emptied only because the Penjualan was stored.
		expect(keranjangState.items).toEqual([]);

		await user.click(struk().getByRole('button', { name: 'Penjualan Baru' }));

		expect(await screen.findByText(/Keranjang kosong/)).toBeInTheDocument();
	});

	it('keeps the keranjang when the API refuses the checkout', async () => {
		serveCatalogue();
		mock.onPost('/penjualan').reply(409, {
			message: 'Stok Kopi Susu tidak cukup: tersisa 1, diminta 2.',
			error: 'insufficient_stock'
		});
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);
		await tambah(user, 'Kopi Susu');
		await tambah(user, 'Kopi Susu');

		await user.type(pembayaran().getByLabelText('Jumlah bayar'), '50000');
		await user.click(pembayaran().getByRole('button', { name: 'Bayar & Simpan Penjualan' }));

		expect(
			await pembayaran().findByText('Stok Kopi Susu tidak cukup: tersisa 1, diminta 2.')
		).toBeInTheDocument();
		// A refused checkout leaves the Kasir's work exactly where it was.
		expect(keranjangState.items).toHaveLength(1);
		expect(keranjangState.items[0]!.qty).toBe(2);
	});

	it('narrows the listing to the Produk a scanned Kode matches', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);

		// What a scanner types into the Kode field: the listing narrows to the one
		// Produk that Kode belongs to, and the Tambah button adds it.
		await user.type(screen.getByLabelText('Kode'), 'KOPI-01');

		expect(
			await screen.findByRole('button', { name: 'Tambah Kopi Susu ke keranjang' })
		).toBeEnabled();
		expect(
			screen.queryByRole('button', { name: 'Tambah Teh Botol ke keranjang' })
		).not.toBeInTheDocument();

		await user.click(screen.getByRole('button', { name: 'Tambah Kopi Susu ke keranjang' }));

		expect(keranjangState.items).toHaveLength(1);
		expect(keranjangState.items[0]!.produk.id).toBe(1);
	});

	it('does not guess which Produk a name search meant', async () => {
		serveCatalogue();
		const user = userEvent.setup();

		renderKasir();
		await screen.findByText(/Kopi Susu/);

		// "o" matches both Kopi Susu and Teh Botol, so both stay on offer.
		await user.type(screen.getByLabelText('Nama'), 'o');

		expect(keranjangState.items).toEqual([]);
		expect(screen.getByRole('button', { name: 'Tambah Kopi Susu ke keranjang' })).toBeEnabled();
		expect(screen.getByRole('button', { name: 'Tambah Teh Botol ke keranjang' })).toBeEnabled();
	});
});
